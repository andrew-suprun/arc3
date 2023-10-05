package main

import (
	"arc/exec"
	"arc/log"
	"bufio"
	"flag"
	"io"
	"os"
	"runtime/debug"
)

var (
	execFlag = flag.String("e", "", "executable")
	outFlag  = flag.String("o", "", "output")
)

func main() {
	flag.Parse()
	if *outFlag == "" {
		log.SetLogger("log.log")
		defer log.CloseLogger()
		log.Debug("provide -o flag")
		return
	}

	log.SetLogger(*outFlag)
	defer log.CloseLogger()

	if *execFlag == "" {
		log.Debug("provide -e flag")
		return
	}

	defer func() {
		if err := recover(); err != nil {
			log.Debug("ERROR", "err", err)
			log.Debug("STACK", "stack", debug.Stack())
		}
	}()

	in, out := exec.Start(*execFlag)

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
