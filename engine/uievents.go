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
		folder.moveOffset(-m.fileTreeLines, m.fileTreeLines)
		folder.moveSelection(-m.fileTreeLines)

	case "PgDn":
		log.Debug("PgDn", "fileTreeLines", m.fileTreeLines)
		folder := m.curArchive.curFolder
		folder.moveOffset(m.fileTreeLines, m.fileTreeLines)
		folder.moveSelection(m.fileTreeLines)

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

func (f *folder) moveOffset(lines, fileTreeLines int) {
	defer func() { log.Debug("moveOffset", "offset.9", f.offsetIdx) }()
	log.Debug("moveOffset", "offset.1", f.offsetIdx)
	f.offsetIdx += lines
	log.Debug("moveOffset", "offset.2", f.offsetIdx)
	entries := len(f.children) + len(f.files)
	log.Debug("moveOffset", "entries", entries)

	if f.offsetIdx >= entries+1-fileTreeLines {
		f.offsetIdx = entries - fileTreeLines
		log.Debug("moveOffset", "offset.3", f.offsetIdx)
	}
	if f.offsetIdx < 0 {
		f.offsetIdx = 0
		log.Debug("moveOffset", "offset.4", f.offsetIdx)
	}
	log.Debug("moveOffset", "offset.8", f.offsetIdx)
}
