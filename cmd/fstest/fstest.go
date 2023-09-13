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

	scanFolder(root, "", archives[root])
	send("archive-scanned", "root", root)
	hashFolder(root, "", archives[root])
	send("archive-hashed", "root", root)
}

func scanFolder(root, path string, folder []*fileMeta) {
	for _, meta := range folder {
		var metaPath string
		if path == "" {
			metaPath = meta.name
		} else {
			metaPath = filepath.Join(path, meta.name)
		}
		if meta.hash != "" {
			send("file-scanned",
				"root", root,
				"path", metaPath,
				"size", meta.size,
				"mod-time", meta.modTime)
		} else {
			send("folder-scanned", "root", root, "path", metaPath)
			scanFolder(root, metaPath, meta.children)
		}
	}
}

func hashFolder(root, path string, folder []*fileMeta) {
	for _, meta := range folder {
		if quit {
			return
		}
		var metaPath string
		if path == "" {
			metaPath = meta.name
		} else {
			metaPath = filepath.Join(path, meta.name)
		}
		if meta.hash != "" {
			if *progress {
				for size := 0; size < meta.size; size += 100000 {
					send("hashing-progress",
						"root", root,
						"path", metaPath,
						"size", size)
					time.Sleep(100 * time.Millisecond)
				}
			}
			send("file-hashed",
				"root", root,
				"path", metaPath,
				"hash", meta.hash)
		} else {
			hashFolder(root, metaPath, meta.children)
		}
	}
}

func send(kind string, params ...any) {
	outgoing <- parser.String(kind, params...)
}

var beginning = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
var end = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
var duration = end.Sub(beginning)

type fileMeta struct {
	name     string
	hash     string
	size     int
	modTime  time.Time
	children []*fileMeta
}

var sizes = map[string]int{}
var modTimes = map[string]time.Time{}

func init() {
	for _, archive := range archives {
		initFolder(archive)
	}
}

func initFolder(folder []*fileMeta) {
	for _, meta := range folder {
		size, ok := sizes[meta.hash]
		if !ok {
			size = rand.Intn(100000000)
			meta.size = size
			sizes[meta.hash] = size
		}
		meta.size = size
		modTime, ok := modTimes[meta.hash]
		if !ok {
			modTime = beginning.Add(time.Duration(rand.Int63n(int64(duration))))
			meta.modTime = modTime
			modTimes[meta.hash] = modTime
		}
		meta.modTime = modTime
		initFolder(meta.children)
	}
}

var archives = map[string][]*fileMeta{
	"origin": origin,
	"copy 1": copy1,
	"copy 2": copy2,
}

var origin = []*fileMeta{
	{name: "a", children: []*fileMeta{
		{name: "b", children: []*fileMeta{
			{name: "c", children: []*fileMeta{
				{name: "d", hash: "0001"},
			}},
		}},
	}},
	{name: "b", hash: "0001"},
	{name: "d", hash: "0002"},
	{name: "e", hash: "0003"},
	{name: "f", hash: "0004"},
	{name: "b1", hash: "0001"},
	{name: "d1", hash: "0002"},
	{name: "e1", hash: "0003"},
	{name: "f1", hash: "0004"},
	{name: "b2", hash: "0001"},
	{name: "d2", hash: "0002"},
	{name: "e2", hash: "0003"},
	{name: "f2", hash: "0004"},
	{name: "b3", hash: "0001"},
	{name: "d3", hash: "0002"},
	{name: "e3", hash: "0003"},
	{name: "f3", hash: "0004"},
	{name: "b4", hash: "0001"},
	{name: "d4", hash: "0002"},
	{name: "e4", hash: "0003"},
	{name: "f4", hash: "0004"},
	{name: "b5", hash: "0001"},
	{name: "d5", hash: "0002"},
	{name: "e5", hash: "0003"},
	{name: "f5", hash: "0004"},
	{name: "b6", hash: "0001"},
	{name: "d6", hash: "0002"},
	{name: "e6", hash: "0003"},
	{name: "f6", hash: "0004"},
	{name: "b7", hash: "0001"},
	{name: "d7", hash: "0002"},
	{name: "e7", hash: "0003"},
	{name: "f7", hash: "0004"},
	{name: "b8", hash: "0001"},
	{name: "d8", hash: "0002"},
	{name: "e8", hash: "0003"},
	{name: "f8", hash: "0004"},
	{name: "b9", hash: "0001"},
	{name: "d9", hash: "0002"},
	{name: "e9", hash: "0003"},
	{name: "f9", hash: "0004"},
	{name: "b10", hash: "0001"},
	{name: "d10", hash: "0002"},
	{name: "e10", hash: "0003"},
	{name: "f10", hash: "0004"},
}
var copy1 = []*fileMeta{
	{name: "a", children: []*fileMeta{
		{name: "b", children: []*fileMeta{
			{name: "c", children: []*fileMeta{
				{name: "d", hash: "0001"},
			}},
		}},
	}},
}
var copy2 = []*fileMeta{
	{name: "a", children: []*fileMeta{
		{name: "b", children: []*fileMeta{
			{name: "c", children: []*fileMeta{
				{name: "d", hash: "0001"},
			}},
		}},
	}},
}
