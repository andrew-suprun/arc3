package engine

import (
	"arc/log"
	"arc/parser"
	"bufio"
	"io"
	"os"
	"runtime/debug"
	"slices"
)

func Run(fsEvents io.ReadCloser, fsCommands io.WriteCloser) {
	defer func() {
		if err := recover(); err != nil {
			log.Debug("ERROR", "err", err)
			log.Debug("STACK", "stack", debug.Stack())
			os.Stdout.WriteString("stopped\n")
		}
	}()

	m := &model{
		archives:    map[string]*archive{},
		filesByHash: map[string][]*meta{},
		fsEvents:    fsEvents,
		fsCommands:  fsCommands,
		uiEvents:    os.Stdin,
		uiCommands:  os.Stdout,
	}

	msgs := make(chan *parser.Message)

	go m.readEvents(fsEvents, msgs)
	go m.readEvents(os.Stdin, msgs)

	for {
		m.handleEvent(<-msgs)
	msgLoop:
		for {
			select {
			case msg := <-msgs:
				m.handleEvent(msg)
			default:
				break msgLoop
			}
		}
		if m.quit {
			break
		}
	}

	for msg := range msgs {
		if msg.Type == "stopped" {
			break
		}
	}
	m.sendToUi("stopped")
}

func (m *model) readEvents(input io.Reader, messages chan *parser.Message) {
	reader := bufio.NewReader(input)

	for !m.quit {
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		messages <- parser.Parse(text)
	}
}

func (m *meta) path() []string {
	if m.parent == nil {
		return nil
	}
	res := []string{}
	m = m.parent
	for m.parent != nil {
		res = append(res, m.name)
		m = m.parent
	}

	slices.Reverse(res)
	return res
}
