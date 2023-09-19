package engine

import (
	"arc/log"
	"arc/parser"
	"arc/renderer"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

func (m *model) handleEvent(msg *parser.Message) {
	switch msg.Type {
	case "current-folder":
		root := msg.StringValue("root")
		path := parsePath(msg.StringValue("path"))
		if root != m.curRoot || slices.Equal(path, m.curPath) {
			m.curRoot = root
			m.curPath = path
			m.sendCurFolder()
		}

	case "scan":
		root := msg.StringValue("root")
		folder := &meta{
			root:     root,
			name:     "",
			children: map[string]*meta{},
		}

		m.archives[root] = &archive{
			root:       root,
			idx:        len(m.roots),
			rootFolder: folder,
			curFolder:  folder,
		}

		m.roots = append(m.roots, root)

		m.sendToFs("scan", "root", root)

	case "folder-scanned":
		root := msg.StringValue("root")
		path, name := parseName(msg.StringValue("path"))
		curFolder := m.folder(root, path)
		curFolder.children[name] = &meta{
			root:     root,
			name:     name,
			parent:   curFolder,
			children: map[string]*meta{},
		}

	case "file-scanned":
		root := msg.StringValue("root")
		path, name := parseName(msg.StringValue("path"))
		size := msg.Int("size")
		modTime := msg.Time("mod-time")

		curFolder := m.folder(root, path)
		curFolder.children[name] = &meta{
			root:    root,
			name:    name,
			parent:  curFolder,
			size:    size,
			modTime: modTime,
			state:   scanned,
		}
		if root == m.curRoot && slices.Equal(path, m.curPath) {
			m.sendToUi("add-file", "kind", "R", "name", name, "size", size, "mod-time", modTime)
		}

	case "archive-scanned":
		root := msg.StringValue("root")
		m.archives[root].state = archiveHashing

	case "hashing-progress", "copying-progress":
		root := msg.StringValue("root")
		path, name := parseName(msg.StringValue("path"))
		curFolder := m.folder(root, path)
		file := curFolder.children[name]
		file.progress = msg.Int("size")

	case "file-hashed":
		root := msg.StringValue("root")
		path, name := parseName(msg.StringValue("path"))
		curFolder := m.folder(root, path)
		hash := msg.StringValue("hash")
		file := curFolder.children[name]
		file.hash = hash
		m.filesByHash[hash] = append(m.filesByHash[hash], file)

	case "archive-hashed":
		root := msg.StringValue("root")
		m.archives[root].state = archiveReady

		for _, archive := range m.archives {
			if archive.state == archiveScanning {
				return
			}
		}
		m.analyzeDiscrepancies()

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

func (m *model) sendCurFolder() {
	m.sendToUi("folder-begin")
	m.sendToUi("folder", "root", m.curRoot, "path", filepath.Join(m.curPath...))

	for _, meta := range m.curArchive().curFolder.children {
		m.sendToUi("meta", "kind", meta.kind, "name", meta.name, "size", meta.size, "mod-time", meta.modTime,
			"state", meta.state, "progress", meta.progress, "counts", counts(meta.counts))
	}

	m.sendToUi("folder-end")
}

func parsePath(strPath string) []string {
	path := strings.Split(string(strPath), "/")
	if path[0] != "" {
		return path
	}
	return nil
}

func parseName(strPath string) ([]string, string) {
	path := strings.Split(string(strPath), "/")
	name := path[len(path)-1]
	return path[:len(path)-1], name
}

func (m *model) folder(root string, path []string) *meta {
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
