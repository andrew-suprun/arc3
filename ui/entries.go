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

type entry struct {
	kind     kind
	name     string
	size     int
	modTime  time.Time
	state    state
	progress int
	counts   string
}

type kind int

const (
	kindRegular kind = iota
	kindFolder
)

func (k kind) String() string {
	switch k {
	case kindRegular:
		return "regular"
	case kindFolder:
		return "folder"
	}
	return ""
}

func parseKind(text string) kind {
	switch text {
	case "regular":
		return kindRegular
	case "folder":
		return kindFolder
	}
	panic("Invalid kind")
}

type state int

const (
	resolved state = iota
	scanned
	pending
	inProgress
	divergent
)

func (s state) String() string {
	switch s {
	case resolved:
		return "resolved"
	case scanned:
		return "scanned"
	case pending:
		return "pending"
	case inProgress:
		return "in-progress"
	case divergent:
		return "divergent"
	}
	return "UNKNOWN FILE STATE"
}

func uiState(engState string) state {
	switch engState {
	case "resolved":
		return resolved
	case "scanned":
		return scanned
	case "in-progress":
		return inProgress
	case "pending":
		return pending
	case "divergent":
		return divergent
	}
	panic("Invalid engine state")
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
