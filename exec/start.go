package exec

import (
	"arc/log"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Start() (in io.ReadCloser, out io.WriteCloser) {
	idx := 0
	for idx = range os.Args {
		if os.Args[idx] == "--" {
			break
		}
	}
	if idx == 0 || idx+1 == len(os.Args) {
		panic(fmt.Sprintf("exec.Start(): invalid parameters: %v", os.Args))
	}

	args := os.Args[idx+1:]

	executable, _ := os.Executable()
	dir := filepath.Dir(executable)
	proc := strings.Replace(args[0], "$", dir, 1)

	cmd := exec.Command(proc, args[1:]...)
	in, _ = cmd.StdoutPipe()
	out, _ = cmd.StdinPipe()
	cmdErr := cmd.Start()
	if cmdErr != nil {
		log.Debug("failed to start", "proc", proc, "error", cmdErr)
		return
	}
	return in, out
}
