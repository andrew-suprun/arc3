package ui

import (
	"arc/exec"
	"arc/log"
	"arc/parser"
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

type app struct {
	view
	events   io.WriteCloser
	commands io.ReadCloser
	incoming chan any
	outgoing chan string
	screen   tcell.Screen
	quit     bool
}

func Run(screen tcell.Screen) {
	commands, events := exec.Start()

	app := &app{
		commands: commands,
		events:   events,
		incoming: make(chan any),
		outgoing: make(chan string),
		screen:   screen,
	}

	screen.EnableMouse()
	go app.sendEvents()
	go app.handleCommands()
	go app.handleTcellEvents()

	app.send("current-folder", "root", os.Args[1], "path", "")

	for _, root := range os.Args[1:] {
		if root == "--" {
			break
		}
		app.send("scan", "root", root)
	}

	app.handleMessages()
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

func (a *app) handleMessages() {
	for message := range a.incoming {
		switch message := message.(type) {
		case *parser.Message:
			a.handleCommand(message)

		case *tcell.EventResize:
			a.sync = true
			a.screenSize.width, a.screenSize.height = message.Size()
			a.send("ready")

		case *tcell.EventKey:
			a.send("key", "name", message.Name())

		case *tcell.EventMouse:
			a.handleMouseEvent(message)

		default:
			panic(fmt.Sprintf("### unhandled renderer event: %T", message))
		}
		if a.quit {
			break
		}
	}
}

func (app *app) handleCommand(command *parser.Message) {
	switch command.Type {
	case "current-folder":
		app.root = command.StringValue("root")
		app.path = command.StringValue("path")
		app.entries.reset()

	case "add-entry":
		app.entries.addEntry(
			kind(command.Int("kind")),
			command.StringValue("name"),
			command.Int("size"),
			command.Time("modTime"),
			state(command.Int("state")),
			command.Int("progress"),
			command.StringValue("counts"),
		)

	case "update-entry":
		app.entries.updateEntry(
			command.StringValue("name"),
			state(command.Int("state")),
			command.Int("progress"),
			command.StringValue("counts"),
		)
		app.render(app.screen)

	case "show-folder":
		app.sort()
		app.render(app.screen)

	case "stopped":
		app.quit = true
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
