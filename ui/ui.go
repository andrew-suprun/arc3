package ui

import (
	"arc/exec"
	"arc/log"
	"arc/parser"
	"arc/renderer"
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

type app struct {
	events   io.WriteCloser
	commands io.ReadCloser
	incoming chan any
	outgoing chan string
	screen   tcell.Screen
	view     view
	// size             size
	// mouseTargetAreas []target
	// scrollAreas      []target
	// position         position
	quit bool
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
			a.view.sync = true
			a.view.screenSize.width, a.view.screenSize.height = message.Size()
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

func (r *app) handleCommand(command *parser.Message) {
	switch command.Type {
	case "stopped":
		r.quit = true
	}
}

func (r *app) handleMouseEvent(event *tcell.EventMouse) {
	x, y := event.Position()

	if event.Buttons() == 256 || event.Buttons() == 512 {
		for _, target := range r.view.scrollAreas {
			if target.position.x <= x && target.position.x+target.size.width > x &&
				target.position.y <= y && target.position.y+target.size.height > y {

				if event.Buttons() == 512 {
					r.send("scroll-up")
				} else {
					r.send("scroll-down")
				}
				return
			}
		}
	}

	for _, target := range r.view.mouseTargetAreas {
		if target.position.x <= x && target.position.x+target.size.width > x &&
			target.position.y <= y && target.position.y+target.size.height > y {

			r.send("mouse-target", "command", target.command)
			return
		}
	}
}

func tcStyle(msg *parser.Message) tcell.Style {
	flags := renderer.Flags(msg.Int("flags"))
	return tcell.StyleDefault.
		Foreground(tcell.PaletteColor(msg.Int("fg"))).
		Background(tcell.PaletteColor(msg.Int("bg"))).
		Bold(flags&renderer.Bold == renderer.Bold).
		Italic(flags&renderer.Italic == renderer.Italic).
		Reverse(flags&renderer.Reverse == renderer.Reverse)
}

func bgStyle(msg *parser.Message) tcell.Style {
	return tcell.StyleDefault.Background(tcell.PaletteColor(msg.Int("bg")))
}

func (r *app) text(msg *parser.Message) {
	// style := tcStyle(msg)
	// for _, char := range msg.StringValue("text") {
	// 	r.screen.SetContent(r.position.x, r.position.y, char, nil, style)
	// 	r.position.x += 1
	// }
}

func (r *app) space(msg *parser.Message) {
	// w := msg.Int("width")
	// h := msg.Int("height")
	// style := bgStyle(msg)

	// for y := 0; y < h; y++ {
	// 	for x := 0; x < w; x++ {
	// 		r.screen.SetContent(r.position.x+x, r.position.y+y, ' ', nil, style)
	// 	}
	// }
}

func (r *app) mouseTarget(msg *parser.Message) {
	// posX := msg.Int("x")
	// posY := msg.Int("y")
	// w := msg.Int("width")
	// h := msg.Int("height")

	// r.mouseTargetAreas = append(r.mouseTargetAreas, target{
	// 	command:  msg.StringValue("command"),
	// 	position: position{x: posX, y: posY},
	// 	size:     size{width: w, height: h},
	// })
}

func (r *app) scroll(msg *parser.Message) {
	// posX := msg.Int("x")
	// posY := msg.Int("y")
	// w := msg.Int("width")
	// h := msg.Int("height")

	// r.scrollAreas = append(r.scrollAreas, target{
	// 	command:  msg.StringValue("command"),
	// 	position: position{x: posX, y: posY},
	// 	size:     size{width: w, height: h},
	// })
}

func (r *app) show() {
	// if r.sync {
	// 	r.screen.Sync()
	// 	r.sync = false
	// } else {
	// 	r.screen.Show()
	// }
	// r.send("ready")
	// r.mouseTargetAreas = r.mouseTargetAreas[:0]
	// r.scrollAreas = r.scrollAreas[:0]
}

func (r *app) send(kind string, params ...any) {
	msg := parser.String(kind, params...)
	r.outgoing <- msg
}
