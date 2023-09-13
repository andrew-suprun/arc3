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
	case "PgDn":
	case "Top":
	case "Bottom":
	}
}

func (f *folder) moveSelection(lines int) {
	f.selectedName = ""
	f.selectedIdx += lines
	entries := len(f.children) + len(f.files)

	if f.selectedIdx >= entries {
		f.selectedIdx = entries - 1
	}
	if f.selectedIdx < 0 {
		f.selectedIdx = 0
	}
}
