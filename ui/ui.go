package ui

import (
	"arc/exec"
	"arc/log"
	"arc/parser"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

type app struct {
	screen              tcell.Screen
	roots               []string
	archives            map[string]*archive
	folders             folders
	root                string
	path                string
	entries             entries
	screenSize          size
	selectFolderTargets []target
	events              io.WriteCloser
	commands            io.ReadCloser
	incoming            chan any
	outgoing            chan string

	folderUpdateInProgress bool
	makeSelectedVisible    bool
	sync                   bool
	quit                   bool
}

type target struct {
	param string
	position
	size
}

type position struct {
	x, y int
}

type size struct {
	width, height int
}

type archive struct {
	state      archiveState
	rootFolder *folder
}

func Run(screen tcell.Screen) {
	commands, events := exec.Start()

	app := &app{
		screen:   screen,
		archives: map[string]*archive{},
		folders:  folders{},
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
		app.send("scan", "root", root)
	}

	app.send("set-current-folder", "root", os.Args[1], "path", "")

	app.handleMessages()
}

func (app *app) reset() {
	app.entries.entries = app.entries.entries[:0]
}

func (app *app) curFolder() *folder {
	return app.folders.folder(app.root, app.path)
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
		app.path = command.StringValue("path")
		app.reset()
		app.folderUpdateInProgress = true

	case "update-entry":
		update := parseEntry(command)
		for i, entry := range app.entries.entries {
			if entry.name == update.name {
				app.entries.entries[i] = update
				return
			}
		}
		app.entries.entries = append(app.entries.entries, update)

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

	case "Down":
		app.curFolder().selectedIdx++

	case "PgUp":
		app.curFolder().selectedIdx -= app.screenSize.height - 4

	case "PgDn":
		app.curFolder().selectedIdx += app.screenSize.height - 4

	case "Right":
		idx := app.curFolder().selectedIdx
		entry := app.entries.entries[idx]
		path := filepath.Join(app.path + entry.name)
		app.send("set-current-folder", "root", app.root, "path", path)

	case "Ctrl+C":
		app.send("stop")
	}
}

func (app *app) handleMouseEvent(event *tcell.EventMouse) {
	x, y := event.Position()
	if event.Buttons() == 256 || event.Buttons() == 512 {
		if y >= 3 && y < app.screenSize.height-1 {
			folder := app.curFolder()
			if event.Buttons() == 512 {
				folder.offsetIdx--
			} else {
				folder.offsetIdx++
			}
		}
		return
	}

	if y == 1 {
		for _, target := range app.selectFolderTargets {
			if target.position.x <= x && target.position.x+target.size.width > x &&
				target.position.y <= y && target.position.y+target.size.height > y {

				app.send("mouse-target", "path", target.param)
				return
			}
		}
	}
}

func (r *app) send(kind string, params ...any) {
	msg := parser.String(kind, params...)
	r.outgoing <- msg
}
