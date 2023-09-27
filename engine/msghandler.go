package engine

import (
	"arc/log"
	"arc/parser"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

func (m *model) handleEvent(msg *parser.Message) {
	switch msg.Type {
	case "set-current-folder":
		root := msg.StringValue("root")
		path := parsePath(msg.StringValue("path"))
		if root != m.curRoot || !slices.Equal(path, m.curPath) {
			m.curRoot = root
			m.curPath = path
			m.sendCurFolder()
		}

	case "scan":
		root := msg.StringValue("root")
		folder := &meta{
			root: root,
			name: "",
		}

		m.archives[root] = &archive{
			root:       root,
			idx:        len(m.roots),
			rootFolder: folder,
			curFolder:  folder,
		}

		m.roots = append(m.roots, root)

		m.sendToFs("scan", "root", root)

	case "file-scanned":
		root := msg.StringValue("root")
		path := msg.StringValue("path")
		name := msg.StringValue("name")
		size := msg.Int("size")
		modTime := msg.Time("mod-time")

		curFolder := m.folder(root, path)
		file := &meta{
			kind:    kindRegular,
			root:    root,
			name:    name,
			parent:  curFolder,
			size:    size,
			modTime: modTime,
			state:   scanned,
		}
		curFolder.addChild(file)
		curFolder.updateState()
		m.updateUiEntry(file)

	case "archive-scanned":
		root := msg.StringValue("root")
		m.archives[root].state = archiveHashing

	case "hashing-progress", "copying-progress":
		root := msg.StringValue("root")
		path := msg.StringValue("path")
		name := msg.StringValue("name")
		curFolder := m.folder(root, path)
		file := curFolder.children[name]
		file.state = inProgress
		file.progress = msg.Int("progress")
		file.parent.updateState()
		m.updateUiEntry(file)

	case "file-hashed":
		root := msg.StringValue("root")
		path := msg.StringValue("path")
		name := msg.StringValue("name")
		hash := msg.StringValue("hash")
		curFolder := m.folder(root, path)
		file := curFolder.children[name]
		file.hash = hash
		file.state = resolved
		file.progress = file.size
		m.filesByHash[hash] = append(m.filesByHash[hash], file)
		file.parent.updateState()
		m.updateUiEntry(file)

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

	case "stop":
		m.sendToFs("stop")

	case "stopped":
		m.sendToUi("stopped")
		m.quit = true

	default:
		log.Debug("UNKNOWN event type", "msg", msg)
		panic(fmt.Sprintf("UNKNOWN event type %q", msg.Type))
	}
}

func (m *model) sendCurFolder() {
	m.sendToUi("current-folder", "root", m.curRoot, "path", filepath.Join(m.curPath...))

	for _, file := range m.curArchive().curFolder.children {
		m.sendEntryToUi(file)
	}

	m.sendToUi("show-folder")
}

func (m *model) updateUiEntry(file *meta) {
	if file.root != m.curRoot {
		return
	}

	path := file.path()
	n := len(path) - len(m.curPath)

	if n < 0 {
		return
	}
	if !slices.Equal(path[:len(m.curPath)], m.curPath) {
		return
	}

	for i := 0; i < n; i++ {
		file = file.parent
	}

	m.sendEntryToUi(file)
}

func (m *model) sendEntryToUi(file *meta) {
	m.sendToUi("update-entry",
		"kind", file.kind.String(),
		"name", file.name,
		"size", file.size,
		"mod-time", file.modTime,
		"state", file.state.String(),
		"progress", file.progress,
		"counts", counts(file.counts))
}

func parsePath(strPath string) []string {
	path := strings.Split(string(strPath), "/")
	if path[0] != "" {
		return path
	}
	return nil
}

func (m *model) folder(root string, path string) *meta {
	segments := parsePath(path)
	folder := m.archives[root].rootFolder
	for _, name := range segments {
		child := folder.children[name]
		if child == nil {
			child = &meta{
				kind:     kindFolder,
				root:     root,
				name:     name,
				parent:   folder,
				children: map[string]*meta{},
			}
			folder.addChild(child)
		}
		folder = child
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
