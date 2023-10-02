package ui

import (
	"arc/exec"
	"arc/log"
	"arc/parser"
	"bufio"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

type app struct {
	screen        tcell.Screen
	roots         []string
	archives      map[string]*archive
	root          string
	entries       entries
	screenSize    size
	folderTargets []folderTarget
	sortTargets   []sortTarget
	events        io.WriteCloser
	commands      io.ReadCloser
	incoming      chan any
	outgoing      chan string

	folderUpdateInProgress bool
	makeSelectedVisible    bool
	sync                   bool
	quit                   bool
}

type folder struct {
	selectedIdx   int
	offsetIdx     int
	sortColumn    sortColumn
	sortAscending []bool
}

type sortColumn int

const (
	sortByName sortColumn = iota
	sortByTime
	sortBySize
)

type folders map[string]*folder

type folderTarget struct {
	param  string
	offset int
	width  int
}

type sortTarget struct {
	sortColumn
	offset int
	width  int
}

type position struct {
	x, y int
}

type size struct {
	width, height int
}

type archive struct {
	path       string
	folders    folders
	state      archiveState
	rootFolder *folder
}

type archiveState int

const (
	archiveScanning archiveState = iota
	archiveHashing
	archiveReady
	archiveCopying
)

func (s archiveState) String() string {
	switch s {
	case archiveScanning:
		return "archiveScanning"
	case archiveHashing:
		return "archiveHashing"
	case archiveReady:
		return "archiveReady"
	case archiveCopying:
		return "archiveCopying"
	}
	panic("Invalid archiveState")
}

func Run(screen tcell.Screen) {
	commands, events := exec.Start()

	app := &app{
		screen:   screen,
		archives: map[string]*archive{},
		commands: commands,
		events:   events,
		incoming: make(chan any),
		outgoing: make(chan string),
	}

	screen.EnableMouse()
	go app.sendEvents()
	go app.handleCommands()
	go app.handleTcellEvents()

	for _, root := range os.Args[1:] {
		if root == "--" {
			break
		}
		app.roots = append(app.roots, root)
		app.archives[root] = &archive{
			folders: folders{},
		}
		app.send("scan", "root", root)
	}

	app.root = os.Args[1]
	app.send("set-current-folder", "root", app.root, "path", "")

	app.handleMessages()
}

func (app *app) reset() {
	app.entries = app.entries[:0]
}

func (app *app) curArchive() *archive {
	return app.archives[app.root]
}

func (app *app) curPath() string {
	return app.curArchive().path
}

func (app *app) folder(root, path string) *folder {
	var f *folder
	var ok bool
	if f, ok = app.archives[root].folders[app.curPath()]; !ok {
		f = &folder{
			sortAscending: []bool{true, true, true},
		}
		app.archives[root].folders[app.curPath()] = f
	}
	return f
}

func (app *app) curFolder() *folder {
	return app.folder(app.root, app.curPath())
}

func (app *app) curEntry() *entry {
	return &app.entries[app.curFolder().selectedIdx]
}

func (r *app) sendEvents() {
	for event := range r.outgoing {
		r.events.Write([]byte(event))
	}
}

func (r *app) handleCommands() {
	defer func() {
		if err := recover(); err != nil {
			log.Debug("ERROR", "err", err)
			log.Debug("STACK", "stack", debug.Stack())
			r.quit = true
		}
	}()

	reader := bufio.NewReader(r.commands)
	for !r.quit {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			r.quit = true
			break
		}
		if err != nil {
			panic(err)
		}
		r.incoming <- parser.Parse(text)
	}
}

func (r *app) handleTcellEvents() {
	defer func() {
		if err := recover(); err != nil {
			log.Debug("ERROR", "err", err)
			log.Debug("STACK", "stack", debug.Stack())
			r.quit = true
		}
	}()

	for !r.quit {
		tcEvent := r.screen.PollEvent()
		for {
			if ev, mouseEvent := tcEvent.(*tcell.EventMouse); !mouseEvent || ev.Buttons() != 0 {
				break
			}
			tcEvent = r.screen.PollEvent()
		}

		if tcEvent != nil {
			r.incoming <- tcEvent
		}
	}
}

func (app *app) handleMessages() {
	for message := range app.incoming {
		switch message := message.(type) {
		case *parser.Message:
			app.handleCommand(message)

		case *tcell.EventResize:
			app.sync = true
			app.screenSize.width, app.screenSize.height = message.Size()
			app.render()

		case *tcell.EventKey:
			app.handleKeyEvent(message)

		case *tcell.EventMouse:
			app.handleMouseEvent(message)

		default:
			panic(fmt.Sprintf("### unhandled renderer event: %T", message))
		}
		if app.quit {
			break
		}
		app.render()
	}
}

func (app *app) handleCommand(command *parser.Message) {
	switch command.Type {
	case "current-folder":
		app.root = command.StringValue("root")
		app.archives[app.root].path = command.StringValue("path")
		app.reset()
		app.folderUpdateInProgress = true

	case "update-entry":
		update := parseEntry(command)
		for i, entry := range app.entries {
			if entry.name == update.name {
				app.entries[i] = update
				return
			}
		}
		app.entries = append(app.entries, update)

	case "show-folder":
		app.sort()
		app.folderUpdateInProgress = false

	case "stopped":
		app.quit = true
	}
}

func parseEntry(msg *parser.Message) entry {
	return entry{
		kind:     parseKind(msg.StringValue("kind")),
		name:     msg.StringValue("name"),
		size:     msg.Int("size"),
		modTime:  msg.Time("mod-time"),
		state:    uiState(msg.StringValue("state")),
		progress: msg.Int("progress"),
		counts:   msg.StringValue("counts"),
	}
}

func (app *app) handleKeyEvent(event *tcell.EventKey) {
	log.Debug("handleKeyEvent", "key", event.Name())
	switch event.Name() {
	case "Up":
		app.curFolder().selectedIdx--
		app.makeSelectedVisible = true

	case "Down":
		app.curFolder().selectedIdx++
		app.makeSelectedVisible = true

	case "PgUp":
		app.curFolder().selectedIdx -= app.screenSize.height - 4
		app.curFolder().offsetIdx -= app.screenSize.height - 4

	case "PgDn":
		app.curFolder().selectedIdx += app.screenSize.height - 4
		app.curFolder().offsetIdx += app.screenSize.height - 4

	case "Home":
		app.curFolder().selectedIdx = 0
		app.makeSelectedVisible = true

	case "End":
		app.curFolder().selectedIdx = len(app.entries) - 1
		app.makeSelectedVisible = true

	case "Right":
		entry := app.curEntry()
		if entry.kind == kindFolder {
			path := filepath.Join(app.curPath(), entry.name)
			app.send("set-current-folder", "root", app.root, "path", path)
		}

	case "Left":
		segments := parsePath(app.curPath())
		if segments != nil {
			path := filepath.Join(segments[:len(segments)-1]...)
			app.send("set-current-folder", "root", app.root, "path", path)
		}

	case "Ctrl+F":
		path := filepath.Join(app.root, app.curPath(), app.curEntry().name)
		osexec.Command("open", "-R", path).Start()

	case "Enter":
		path := filepath.Join(app.root, app.curPath(), app.curEntry().name)
		osexec.Command("open", path).Start()

	case "Ctrl+C":
		app.send("stop")

	case "Ctrl+R":
		// TODO Resole

	case "Ctrl+A":
		// TODO Resole All

	case "Tab":
		// TODO Tab

	case "Backspace2": // Ctrl+Delete
		// TODO Delete

	case "F10":
		// TODO Switch Debug On/Off

	case "F12":
		// TODO Print App State

	default:
		if event.Name() >= "Rune[1]" && event.Name() <= "Rune[9]" {
			arcIdx := int(event.Name()[5] - '1')
			if arcIdx < len(app.roots) {
				root := app.roots[arcIdx]
				path := app.archives[root].path
				app.send("set-current-folder", "root", root, "path", path)
			}
		}
	}
}

func (app *app) handleMouseEvent(event *tcell.EventMouse) {
	x, y := event.Position()
	if event.Buttons() == 256 || event.Buttons() == 512 {
		if y >= 3 && y < app.screenSize.height-1 {
			folder := app.curFolder()
			if event.Buttons() == 512 {
				folder.offsetIdx++
			} else {
				folder.offsetIdx--
			}
		}
		return
	}

	if y == 1 {
		for _, target := range app.folderTargets {
			if target.offset <= x && target.offset+target.width > x {
				app.send("set-current-folder", "root", app.root, "path", target.param)
				return
			}
		}
	} else if y == 2 {
		for i, target := range app.sortTargets {
			if target.offset <= x && x < target.offset+target.width {
				folder := app.curFolder()
				if folder.sortColumn == target.sortColumn {
					folder.sortAscending[i] = !folder.sortAscending[i]
				} else {
					folder.sortColumn = target.sortColumn
				}
				app.sort()
			}
		}
	} else if y >= 3 && y < app.screenSize.height-1 {
		folder := app.curFolder()
		idx := folder.offsetIdx + y - 3
		if idx < len(app.entries) {
			folder.selectedIdx = folder.offsetIdx + y - 3
		}
	}
}

func (r *app) send(kind string, params ...any) {
	msg := parser.String(kind, params...)
	r.outgoing <- msg
}

func (app *app) sort() {
	folder := app.curFolder()
	switch folder.sortColumn {
	case sortByName:
		app.entries.sortByName()
	case sortByTime:
		app.entries.sortByTime()
	case sortBySize:
		app.entries.sortBySize()
	}
	if !folder.sortAscending[folder.sortColumn] {
		app.entries.reverse()
	}
}
