package ui

import (
	"time"
)

func (app *app) curFolder() *folder {
	return app.folders.folder(app.root, app.path)
}

type folder struct {
	selectedIdx   int
	offsetIdx     int
	sortColumn    sortColumn
	sortAscending []bool
}

type folders map[string]map[string]*folder

func (f folders) folder(root, path string) *folder {
	var archiveFolders map[string]*folder
	var curFolder *folder
	var ok bool
	if archiveFolders, ok = f[root]; !ok {
		archiveFolders = map[string]*folder{}
		f[root] = archiveFolders
	}
	if curFolder, ok = archiveFolders[path]; !ok {
		curFolder = &folder{
			sortAscending: []bool{true, true, true},
		}
		archiveFolders[path] = curFolder
	}
	return curFolder
}

func (app *app) sort() {
	folder := app.curFolder()
	switch folder.sortColumn {
	case sortByName:
		app.entries.sortByName()
	case sortBySize:
		app.entries.sortBySize()
	case sortByTime:
		app.entries.sortByTime()
	}
	if !folder.sortAscending[folder.sortColumn] {
		app.entries.reverse()
	}
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

type target struct {
	param string
	position
	size
}

type position struct {
	x, y int
}

type size struct {
	width, height int
}

type kind int

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

const (
	kindRegular kind = iota
	kindFolder
)

type archiveState int

const (
	archiveScanning archiveState = iota
	archiveHashing
	archiveReady
	archiveCopying
)

func (s archiveState) String() string {
	switch s {
	case archiveScanning:
		return "archiveScanning"
	case archiveHashing:
		return "archiveHashing"
	case archiveReady:
		return "archiveReady"
	case archiveCopying:
		return "archiveCopying"
	}
	panic("Invalid archiveState")
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
	case "hashing", "copying":
		return inProgress
	case "pending":
		return pending
	case "divergent":
		return divergent
	}
	panic("Invalid engine state")
}

type sortColumn int

const (
	sortByName sortColumn = iota
	sortBySize
	sortByTime
)
