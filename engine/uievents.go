package engine

import "arc/log"

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

	case "Ctrl+C":
		m.sendToFs("stop")
	}
}

func (f *folder) moveSelection(lines int) {
	f.selectedName = ""
	f.selectedIdx += lines
}
