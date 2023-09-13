package engine

import (
	"arc/log"
	"arc/parser"
	"arc/renderer"
	"fmt"
	"io"
	"strings"
)

func (m *model) handleEvent(msg *parser.Message) {
	if _, ok := nonUpdatingEvents[msg.Type]; !ok {
		m.curArchive.updated = true
	}
	switch msg.Type {
	case "scan":
		root := msg.StringValue("root")
		folder := &folder{
			meta: meta{
				root: root,
				name: "",
			},
			children:      map[string]*folder{},
			files:         map[string]*file{},
			sortAscending: []bool{true, true, true},
		}

		m.archives[root] = &archive{
			root:       root,
			idx:        len(m.roots),
			rootFolder: folder,
			curFolder:  folder,
		}

		m.roots = append(m.roots, root)

		if len(m.roots) == 1 {
			m.curArchive = m.archives[m.roots[0]]
		}

		m.sendToFs("scan", "root", root)

	case "folder-scanned":
		root := msg.StringValue("root")
		path, name := parseName(msg.StringValue("path"))
		curFolder := m.folder(root, path)
		curFolder.children[name] = &folder{
			meta: meta{
				root:   root,
				name:   name,
				parent: curFolder,
			},
			children:      map[string]*folder{},
			files:         map[string]*file{},
			sortAscending: []bool{true, true, true},
		}

	case "file-scanned":
		root := msg.StringValue("root")
		path, name := parseName(msg.StringValue("path"))

		curFolder := m.folder(root, path)
		curFolder.files[name] = &file{
			meta: meta{
				root:    root,
				name:    name,
				parent:  curFolder,
				size:    msg.Int("size"),
				modTime: msg.Time("mod-time"),
				state:   scanned,
			},
		}

	case "archive-scanned":
		root := msg.StringValue("root")
		m.archives[root].state = archiveHashing

	case "hashing-progress", "copying-progress":
		root := msg.StringValue("root")
		path, name := parseName(msg.StringValue("path"))
		curFolder := m.folder(root, path)
		file := curFolder.files[name]
		file.progress = msg.Int("size")

	case "file-hashed":
		root := msg.StringValue("root")
		path, name := parseName(msg.StringValue("path"))
		curFolder := m.folder(root, path)
		curFolder.files[name].hash = msg.StringValue("hash")

	case "archive-hashed":
		root := msg.StringValue("root")
		m.archives[root].state = archiveReady

	case "file-copied":
		panic("IMPLEMENT file-copied")

	case "file-moved":
		panic("IMPLEMENT file-moved")

	case "file-deleted":
		panic("IMPLEMENT file-deleted")

	case "ready":
		m.requestFrame = true

	case "screen-size":
		m.screenSize = renderer.Size{Width: msg.Int("width"), Height: msg.Int("height")}

	case "stop":
		m.sendToFs("stop")

	case "stopped":
		m.sendToUi("stopped")
		m.quit = true

	case "key":
		m.handleKey(msg.StringValue("name"))

	default:
		log.Debug("UNKNOWN event type", "msg", msg)
		panic(fmt.Sprintf("UNKNOWN event type %q", msg.Type))
	}
}

func parseName(strPath string) ([]string, string) {
	path := strings.Split(string(strPath), "/")
	name := path[len(path)-1]
	return path[:len(path)-1], name
}

func (m *model) folder(root string, path []string) *folder {
	folder := m.archives[root].rootFolder
	for _, name := range path {
		folder = folder.children[name]
	}
	return folder
}

func (m *model) sendToFs(kind string, params ...any) {
	m.send(m.fsCommands, kind, params...)
}

func (m *model) sendToUi(kind string, params ...any) {
	m.send(m.uiCommands, kind, params...)
}

func (m *model) send(out io.Writer, kind string, params ...any) {
	out.Write([]byte(parser.String(kind, params...)))
}

var nonUpdatingEvents map[string]struct{} = map[string]struct{}{
	"scan":  {},
	"ready": {},
}
