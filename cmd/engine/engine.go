package main

import (
	"arc/engine"
	"arc/exec"
	"arc/log"
	"os"
	"path/filepath"
)

func main() {
	log.SetLogger("log-engine.log")
	defer log.CloseLogger()

	fs := os.Getenv("ARC_FS")
	if fs == "" {
		executable, _ := os.Executable()
		dir := filepath.Dir(executable)
		fs = filepath.Join(dir, "fs")
	}
	engine.Run(exec.Start(fs))
}
