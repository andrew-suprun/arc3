package tcell

import (
	"arc/message"
	"arc/renderer"
	"log"

	"github.com/gdamore/tcell/v2"
)

type target struct {
	Command any
	renderer.Position
	renderer.Size
}

type tcellRenderer struct {
	messages         chan<- any
	screen           tcell.Screen
	mouseTargetAreas []target
	scrollAreas      []target
	sync             bool
}

func NewRenderer(events chan<- any) (renderer.UI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	screen.EnableMouse()

	renderer := &tcellRenderer{
		messages: events,
		screen:   screen,
	}
	go renderer.handleTcellEvents(events)

	return renderer, nil
}

func (r *tcellRenderer) tcStyle(style renderer.Style) tcell.Style {
	flags := style.Flags
	return tcell.StyleDefault.
		Foreground(tcell.PaletteColor(int(style.FG))).
		Background(tcell.PaletteColor(int(style.BG))).
		Bold(flags&renderer.Bold == renderer.Bold).
		Italic(flags&renderer.Italic == renderer.Italic).
		Reverse(flags&renderer.Reverse == renderer.Reverse)
}

func (r *tcellRenderer) Text(text string, pos renderer.Position, style renderer.Style) {
	tcStyle := r.tcStyle(style)
	for x, char := range []rune(text) {
		r.screen.SetContent(pos.X+x, pos.Y, char, nil, tcStyle)
	}
}

func (r *tcellRenderer) Space(pos renderer.Position, size renderer.Size, style renderer.Style) {
	tcStyle := r.tcStyle(style)
	for y := 0; y < size.Height; y++ {
		for x := 0; y < size.Width; x++ {
			r.screen.SetContent(x, y, ' ', nil, tcStyle)
		}
	}
}

func (r *tcellRenderer) MouseTarget(command any, pos renderer.Position, size renderer.Size) {
	r.mouseTargetAreas = append(r.mouseTargetAreas, target{
		Command:  command,
		Position: pos,
		Size:     size,
	})
}

func (r *tcellRenderer) Scroll(command any, pos renderer.Position, size renderer.Size) {
	r.scrollAreas = append(r.scrollAreas, target{
		Command:  command,
		Position: pos,
		Size:     size,
	})
}

func (r *tcellRenderer) Render() {
	if r.sync {
		r.screen.Sync()
		r.sync = false
	} else {
		r.screen.Show()
	}
}

func (r *tcellRenderer) handleTcellEvents(events chan<- any) {
	for {
		tcEvent := r.screen.PollEvent()
		for {
			if ev, mouseEvent := tcEvent.(*tcell.EventMouse); !mouseEvent || ev.Buttons() != 0 {
				break
			}
			tcEvent = r.screen.PollEvent()
		}

		if tcEvent != nil {
			r.handleTcellEvent(tcEvent)

		}
	}
}

func (r *tcellRenderer) handleTcellEvent(tcEvent tcell.Event) bool {
	switch tcEvent := tcEvent.(type) {
	case *tcell.EventResize:
		r.sync = true
		x, y := tcEvent.Size()
		r.messages <- message.ScreenSize{Width: x, Height: y}

	case *tcell.EventMouse:
		r.handleMouseEvent(tcEvent)

	case *tcell.EventKey:
		r.handleKeyEvent(tcEvent)

	default:
		log.Panicf("### unhandled renderer event: %T", tcEvent)
	}
	return true
}

func (r *tcellRenderer) handleKeyEvent(key *tcell.EventKey) {
	log.Printf("### key: %q", key.Name())
	switch key.Name() {
	case "Ctrl+C":
		r.messages <- message.Quit{}

	case "Enter":
		r.messages <- message.Open{}

	// case "Esc":

	case "Ctrl+F":
		r.messages <- message.RevealInFinder{}

	case "Home":
		r.messages <- message.SelectFirst{}

	case "End":
		r.messages <- message.SelectLast{}

	case "PgUp":
		r.messages <- message.PgUp{}

	case "PgDn":
		r.messages <- message.PgDn{}

	case "Up":
		r.messages <- message.MoveSelection{Lines: -1}

	case "Down":
		r.messages <- message.MoveSelection{Lines: 1}

	case "Left":
		r.messages <- message.Exit{}

	case "Right":
		r.messages <- message.Enter{}

	case "Ctrl+R":
		r.messages <- message.ResolveOne{}

	case "Ctrl+A":
		r.messages <- message.ResolveAll{}

	// case "Ctrl+A":
	// 	device.events <- message.KeepAll{}

	case "Tab":
		r.messages <- message.Tab{}

	case "Backspace2": // Ctrl+Delete
		r.messages <- message.Delete{}

	case "F10":
		r.messages <- message.DebugPrintState{}
	case "F12":
		r.messages <- message.DebugPrintRootWidget{}

	default:
		if key.Name() >= "Rune[1]" && key.Name() <= "Rune[9]" {
			r.messages <- message.SelectArchive{Idx: int(key.Name()[5] - '1')}
		}
	}
}

func (d *tcellRenderer) handleMouseEvent(event *tcell.EventMouse) {
	x, y := event.Position()

	if event.Buttons() == 256 || event.Buttons() == 512 {
		for _, target := range d.scrollAreas {
			if target.Position.X <= x && target.Position.X+target.Size.Width > x &&
				target.Position.Y <= y && target.Position.Y+target.Size.Height > y {

				if event.Buttons() == 512 {
					d.messages <- message.Scroll{Command: target.Command, Lines: 1}
				} else {
					d.messages <- message.Scroll{Command: target.Command, Lines: -1}
				}
				return
			}
		}
	}

	for _, target := range d.mouseTargetAreas {
		if target.Position.X <= x && target.Position.X+target.Size.Width > x &&
			target.Position.Y <= y && target.Position.Y+target.Size.Height > y {

			d.messages <- message.MouseTarget{Command: target.Command}
			return
		}
	}
}
