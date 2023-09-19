package main

import (
	"arc/log"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

func main() {
	log.SetLogger("uitest.log")
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
	run(screen)
}

func run(screen tcell.Screen) {
	screen.EnableMouse()

}
