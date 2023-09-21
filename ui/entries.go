package ui

import (
	"cmp"
	"slices"
	"strings"
	"time"
)

type entries struct {
	entries []entry
}

func (e entries) reset() {
	e.entries = e.entries[:0]
}

func (e entries) addEntry(kind kind, name string, size int, modTime time.Time, state state, progress int, counts string) {
	e.entries = append(e.entries, entry{
		kind:     kind,
		name:     name,
		size:     size,
		modTime:  modTime,
		state:    state,
		progress: progress,
		counts:   counts,
	})
}

func (e entries) updateEntry(name string, state state, progress int, counts string) {
	for i := range e.entries {
		entry := &e.entries[i]
		if entry.name == name {
			entry.state = state
			entry.progress = progress
			entry.counts = counts
			return
		}
	}
}

func (e entries) sortByName() {
	slices.SortFunc(e.entries, func(i, j entry) int {
		byName := cmp.Compare(strings.ToLower(i.name), strings.ToLower(j.name))
		if byName != 0 {
			return byName
		}
		byTime := cmp.Compare(i.size, j.size)
		if byTime != 0 {
			return byTime
		}
		return i.modTime.Compare(j.modTime)
	})
}

func (e entries) sortBySize() {
	slices.SortFunc(e.entries, func(i, j entry) int {
		bySize := cmp.Compare(i.size, j.size)
		if bySize != 0 {
			return bySize
		}
		byName := cmp.Compare(strings.ToLower(i.name), strings.ToLower(j.name))
		if byName != 0 {
			return byName
		}
		return i.modTime.Compare(j.modTime)
	})
}

func (e entries) sortByTime() {
	slices.SortFunc(e.entries, func(i, j entry) int {
		byTime := i.modTime.Compare(j.modTime)
		if byTime != 0 {
			return byTime
		}
		byName := cmp.Compare(strings.ToLower(i.name), strings.ToLower(j.name))
		if byName != 0 {
			return byName
		}
		return cmp.Compare(i.size, j.size)
	})
}

func (e entries) reverse() {
	slices.Reverse(e.entries)
}
