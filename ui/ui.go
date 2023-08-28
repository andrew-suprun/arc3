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

type tcellRenderer struct {
	events           io.WriteCloser
	commands         io.ReadCloser
	incoming         chan any
	outgoing         chan string
	screen           tcell.Screen
	size             size
	mouseTargetAreas []target
	scrollAreas      []target
	position         position
	sync             bool
	quit             bool
}

type target struct {
	command string
	position
	size
}

type position struct {
	x, y int
}

type size struct {
	width, height int
}

func Run(screen tcell.Screen) {
	defer func() {
		if err := recover(); err != nil {
			log.Debug("ERROR", "err", err)
			log.Debug("STACK", "stack", debug.Stack())
		}
	}()

	commands, events := exec.Start()

	renderer := &tcellRenderer{
		commands: commands,
		events:   events,
		incoming: make(chan any),
		outgoing: make(chan string),
		screen:   screen,
	}

	screen.EnableMouse()
	go renderer.sendEvents()
	go renderer.handleCommands()
	go renderer.handleTcellEvents()

	for _, root := range os.Args[1:] {
		if root == "--" {
			break
		}
		renderer.send("scan", "root", root)
	}

	renderer.handleMessages()
}

func (r *tcellRenderer) sendEvents() {
	for event := range r.outgoing {
		r.events.Write([]byte(event))
	}
}

func (r *tcellRenderer) handleCommands() {
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

func (r *tcellRenderer) handleTcellEvents() {
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

func (r *tcellRenderer) handleMessages() {
	for message := range r.incoming {
		switch message := message.(type) {
		case *parser.Message:
			r.handleCommand(message)

		case *tcell.EventResize:
			r.sync = true
			r.size.width, r.size.height = message.Size()
			r.send("screen-size", "width", r.size.width, "height", r.size.height)
			r.send("ready")

		case *tcell.EventMouse:
			r.handleMouseEvent(message)

		case *tcell.EventKey:
			r.handleKeyEvent(message)

		default:
			panic(fmt.Sprintf("### unhandled renderer event: %T", message))
		}
		if r.quit {
			break
		}
	}
}

func (r *tcellRenderer) handleCommand(command *parser.Message) {
	switch command.Type {
	case "pos":
		r.pos(command)
	case "text":
		r.text(command)
	case "space":
		r.space(command)
	case "mouse-target":
		r.mouseTarget(command)
	case "scroll":
		r.scroll(command)
	case "show":
		r.show()
	case "stopped":
		r.quit = true
	}
}

func (r *tcellRenderer) handleKeyEvent(key *tcell.EventKey) {
	name := key.Name()
	if name == "Ctrl+C" {
		r.send("stop")
		return
	}
	r.send("key", "name", name)
}

func (r *tcellRenderer) handleMouseEvent(event *tcell.EventMouse) {
	x, y := event.Position()

	if event.Buttons() == 256 || event.Buttons() == 512 {
		for _, target := range r.scrollAreas {
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

	for _, target := range r.mouseTargetAreas {
		if target.position.x <= x && target.position.x+target.size.width > x &&
			target.position.y <= y && target.position.y+target.size.height > y {

			r.send("mouse-target", "command", target.command)
			return
		}
	}
}

func (r *tcellRenderer) pos(msg *parser.Message) {
	r.position = position{x: msg.Int("x"), y: msg.Int("y")}
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

func (r *tcellRenderer) text(msg *parser.Message) {
	style := tcStyle(msg)
	for _, char := range msg.StringValue("text") {
		r.screen.SetContent(r.position.x, r.position.y, char, nil, style)
		r.position.x += 1
	}
}

func (r *tcellRenderer) space(msg *parser.Message) {
	w := msg.Int("width")
	h := msg.Int("height")
	style := bgStyle(msg)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r.screen.SetContent(r.position.x+x, r.position.y+y, ' ', nil, style)
		}
	}
}

func (r *tcellRenderer) mouseTarget(msg *parser.Message) {
	posX := msg.Int("x")
	posY := msg.Int("y")
	w := msg.Int("width")
	h := msg.Int("height")

	r.mouseTargetAreas = append(r.mouseTargetAreas, target{
		command:  msg.StringValue("command"),
		position: position{x: posX, y: posY},
		size:     size{width: w, height: h},
	})
}

func (r *tcellRenderer) scroll(msg *parser.Message) {
	posX := msg.Int("x")
	posY := msg.Int("y")
	w := msg.Int("width")
	h := msg.Int("height")

	r.scrollAreas = append(r.scrollAreas, target{
		command:  msg.StringValue("command"),
		position: position{x: posX, y: posY},
		size:     size{width: w, height: h},
	})
}

func (r *tcellRenderer) show() {
	if r.sync {
		r.screen.Sync()
		r.sync = false
	} else {
		r.screen.Show()
	}
	r.send("ready")
	r.mouseTargetAreas = r.mouseTargetAreas[:0]
	r.scrollAreas = r.scrollAreas[:0]
}

func (r *tcellRenderer) send(kind string, params ...any) {
	msg := parser.String(kind, params...)
	r.outgoing <- msg
}
