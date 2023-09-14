package engine

import (
	"time"
)

func (m *model) updateMetas(folder *folder) {
	folder.size = 0
	folder.modTime = time.Time{}
	folder.state = resolved

	for _, child := range folder.children {
		m.updateMetas(child)
		updateMeta(folder, &child.meta)
	}

	for _, file := range folder.files {
		updateMeta(folder, &file.meta)
	}
}

func updateMeta(folder *folder, meta *meta) {
	folder.progress += meta.progress
	folder.size += meta.size
	if folder.modTime.Before(meta.modTime) {
		folder.modTime = meta.modTime
	}
	folder.state = max(folder.state, meta.state)
}
