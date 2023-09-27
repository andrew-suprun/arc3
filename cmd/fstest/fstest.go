package main

import (
	"arc/log"
	"arc/parser"
	"bufio"
	"cmp"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strconv"
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
			for progress := 0; progress < file.size; progress += 1000000 {
				if quit {
					return
				}
				send("hashing-progress",
					"root", root,
					"path", path,
					"name", name,
					"progress", progress)
				// time.Sleep(20 * time.Millisecond)
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

type fileMeta struct {
	name    string
	hash    string
	size    int
	modTime time.Time
}

var archives = map[string][]fileMeta{}

func init() {
	or := readMeta()
	c1 := slices.Clone(or)
	c2 := slices.Clone(or)
	archives = map[string][]fileMeta{
		"origin": or,
		"copy 1": c1,
		"copy 2": c2,
	}
}

func readMeta() []fileMeta {
	result := []fileMeta{}
	hashInfoFile, err := os.Open("data/.meta.csv")
	if err != nil {
		return nil
	}
	defer hashInfoFile.Close()

	records, err := csv.NewReader(hashInfoFile).ReadAll()
	if err != nil || len(records) == 0 {
		return nil
	}

	for _, record := range records[1:] {
		if len(record) == 5 {
			name := record[1]
			size, er2 := strconv.ParseUint(record[2], 10, 64)
			modTime, er3 := time.Parse(time.RFC3339, record[3])
			modTime = modTime.UTC().Round(time.Second)
			hash := record[4]
			if hash == "" || er2 != nil || er3 != nil {
				continue
			}

			result = append(result, fileMeta{
				name:    name,
				hash:    hash,
				size:    int(size),
				modTime: modTime,
			})
		}
	}
	slices.SortFunc(result, func(a, b fileMeta) int {
		return cmp.Compare(a.name, b.name)
	})
	return result
}
