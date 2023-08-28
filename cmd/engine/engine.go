package main

import (
	"arc/engine"
	"arc/exec"
	"arc/log"
)

func main() {
	log.SetLogger("log-engine.log")
	defer log.CloseLogger()

	engine.Run(exec.Start())
}
