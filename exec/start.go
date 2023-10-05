package exec

import (
	"arc/log"
	"io"
	"os/exec"
	"strings"
)

func Start(commandLine string) (in io.ReadCloser, out io.WriteCloser) {
	command := strings.Split(commandLine, " ")
	cmd := exec.Command(command[0], command[1:]...)
	in, _ = cmd.StdoutPipe()
	out, _ = cmd.StdinPipe()
	cmdErr := cmd.Start()
	if cmdErr != nil {
		log.Debug("failed to start", "command", commandLine, "error", cmdErr)
		return
	}
	return in, out
}
