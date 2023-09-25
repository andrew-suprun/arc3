package main

import (
	"arc/log"
	"arc/parser"
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

var progress = flag.Bool("p", true, "show progress")

var wg = &sync.WaitGroup{}
var outgoing = make(chan string)
var quit = false

func main() {
	log.SetLogger("log-fstest.log")
	defer log.CloseLogger()

	defer func() {
		if err := recover(); err != nil {
			log.Debug("ERROR", "err", err)
			log.Debug("STACK", "stack", debug.Stack())
		}
	}()

	flag.Parse()

	go func() {
		for message := range outgoing {
			fmt.Print(message)
		}
	}()

	reader := bufio.NewReader(os.Stdin)

mainLoop:
	for {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		cmd := parser.Parse(text)
		switch cmd.Type {
		case "scan":
			wg.Add(1)
			go scanArchive(cmd.StringValue("root"))
		case "stop":
			quit = true
			wg.Wait()
			break mainLoop
		default:
			log.Debug("unrecognized", "type", cmd.Type)
		}
	}

	fmt.Println("stopped")
}

func scanArchive(root string) {
	defer wg.Done()

	for _, file := range archives[root] {
		path := filepath.Dir(file.name)
		name := filepath.Base(file.name)

		send("file-scanned",
			"root", root,
			"path", path,
			"name", name,
			"size", file.size,
			"mod-time", file.modTime)
	}

	send("archive-scanned", "root", root)

	for _, file := range archives[root] {
		if quit {
			return
		}
		path := filepath.Dir(file.name)
		name := filepath.Base(file.name)

		if *progress {
			for size := 0; size < file.size; size += 100000 {
				if quit {
					return
				}
				send("hashing-progress",
					"root", root,
					"path", path,
					"name", name,
					"size", size)
				time.Sleep(100 * time.Millisecond)
			}
		}
		send("file-hashed",
			"root", root,
			"path", path,
			"name", name,
			"hash", file.hash)
	}

	send("archive-hashed", "root", root)
}

func send(kind string, params ...any) {
	outgoing <- parser.String(kind, params...)
}

var beginning = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
var end = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
var duration = end.Sub(beginning)

type fileMeta struct {
	name    string
	hash    string
	size    int
	modTime time.Time
}

var sizes = map[string]int{}
var modTimes = map[string]time.Time{}

func init() {
	for _, archive := range archives {
		for _, file := range archive {
			size, ok := sizes[file.hash]
			if !ok {
				size = rand.Intn(100000000)
				file.size = size
				sizes[file.hash] = size
			}
			file.size = size
			modTime, ok := modTimes[file.hash]
			if !ok {
				modTime = beginning.Add(time.Duration(rand.Int63n(int64(duration))))
				file.modTime = modTime
				modTimes[file.hash] = modTime
			}
			file.modTime = modTime
		}
	}
}

var archives = map[string][]*fileMeta{
	"origin": {
		{name: "a/b/c/d", hash: "0001"},
	},
	"copy 1": {
		{name: "a/b/c/d", hash: "0001"},
	},
	"copy 2": {
		{name: "a/b/c/d", hash: "0002"},
	},
}
