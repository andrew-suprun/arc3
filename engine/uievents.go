package engine

import (
	"arc/log"
	"os/exec"
	"path/filepath"
)

func (m *model) handleKey(name string) {
	log.Debug("key", "name", name)
	switch name {
	case "Up":
		folder := m.curArchive.curFolder
		folder.moveSelection(-1)
		m.makeSelectedVisible = true

	case "Down":
		folder := m.curArchive.curFolder
		folder.moveSelection(1)
		m.makeSelectedVisible = true

	case "PgUp":
		folder := m.curArchive.curFolder
		folder.offsetIdx -= m.fileTreeLines
		folder.moveSelection(-m.fileTreeLines)

	case "PgDn":
		folder := m.curArchive.curFolder
		folder.offsetIdx += m.fileTreeLines
		folder.moveSelection(m.fileTreeLines)

	case "Home":
		folder := m.curArchive.curFolder
		folder.selectedName = ""
		folder.selectedIdx = 0
		m.makeSelectedVisible = true

	case "End":
		folder := m.curArchive.curFolder
		folder.selectedName = ""
		folder.selectedIdx = len(folder.children) + len(folder.files) - 1
		m.makeSelectedVisible = true

	case "Left":
		archive := m.curArchive
		parent := archive.curFolder.parent
		if parent != nil {
			archive.curFolder = parent
		}

	case "Right":
		archive := m.curArchive
		folder := archive.curFolder
		child := folder.children[folder.selectedName]
		if child != nil {
			archive.curFolder = child
		}

	case "Enter":
		archive := m.curArchive
		folder := archive.curFolder
		pathSegments := folder.path()
		path := filepath.Join(pathSegments...)
		fileName := filepath.Join(archive.root, path, folder.selectedName)
		exec.Command("open", fileName).Start()

	case "Ctrl+F":
		archive := m.curArchive
		folder := archive.curFolder
		path := filepath.Join(folder.path()...)
		fileName := filepath.Join(archive.root, path, folder.selectedName)
		exec.Command("open", "-R", fileName).Start()

	case "Ctrl+C":
		m.sendToFs("stop")
	}
}

func (f *folder) moveSelection(lines int) {
	f.selectedName = ""
	f.selectedIdx += lines
}
