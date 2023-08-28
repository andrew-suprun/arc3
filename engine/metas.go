package engine

import (
	"arc/log"
	"time"
)

func (c *model) updateMetas(folder *folder) {
	log.Debug("updateMetas", "name", folder.name)
	folder.size = 0
	folder.modTime = time.Time{}
	folder.state = resolved

	for _, childIdx := range folder.children {
		c.updateMetas(childIdx)
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
