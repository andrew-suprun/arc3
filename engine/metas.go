package engine

import (
	"time"
)

func (m *model) updateMetas(folder *meta) {
	folder.size = 0
	folder.modTime = time.Time{}
	folder.state = resolved

	for _, child := range folder.children {
		if child.kind == kindFolder {
			m.updateMetas(child)
			updateMeta(folder, child)
		} else {
			updateMeta(folder, child)
		}
	}
}

func updateMeta(folder *meta, meta *meta) {
	folder.progress += meta.progress
	folder.size += meta.size
	if folder.modTime.Before(meta.modTime) {
		folder.modTime = meta.modTime
	}
	folder.state = max(folder.state, meta.state)
}
