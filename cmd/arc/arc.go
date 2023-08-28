package main

import (
	"arc/log"
	"arc/ui"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

func main() {
	log.SetLogger("log-arc.log")
	defer log.CloseLogger()

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Debug("ERROR", err)
		return
	}
	if err := screen.Init(); err != nil {
		log.Debug("ERROR", err)
		return
	}

	defer func() {
		screen.Fini()
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Debug("ERROR", "err", err)
			log.Debug("STACK", "stack", debug.Stack())
		}
	}()

	ui.Run(screen)
}
