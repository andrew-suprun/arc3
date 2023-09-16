package main

import (
	"arc/exec"
	"arc/log"
	"bufio"
	"io"
	"os"
	"runtime/debug"
)

func main() {
	log.SetLogger(os.Args[1])
	defer log.CloseLogger()

	defer func() {
		if err := recover(); err != nil {
			log.Debug("ERROR", "err", err)
			log.Debug("STACK", "stack", debug.Stack())
		}
	}()

	in, out := exec.Start()

	go func() {
		reader := bufio.NewReader(in)

		for {
			text, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			log.Debug("<<<", text)

			_, err = os.Stdout.WriteString(text)
			if err != nil {
				panic(err)
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		log.Debug(">>>", text)

		_, err = out.Write([]byte(text))
		if err != nil {
			panic(err)
		}
	}
}
