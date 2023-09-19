package ui

import (
	"time"
)

type view struct {
	roots               []string
	archives            map[string]*archive
	folders             folders
	root                string
	path                string
	folder              *folder
	entries             []entry
	screenSize          size
	selectFolderTargets []target
	makeSelectedVisible bool
	sync                bool
}

func (v *view) curFolder() *folder {
	return v.folders.folder(v.root, v.path)
}

type archive struct {
	state      archiveState
	rootFolder *folder
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
	if curFolder, ok := archiveFolders[path]; !ok {
		curFolder = &folder{
			sortAscending: []bool{true, true, true},
		}
		archiveFolders[path] = curFolder
	}
	return curFolder
}

type entry struct {
	kind     kind
	name     string
	size     int
	modTime  time.Time
	state    state
	progress int
	counts   []int
	selected bool
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
		return "Regular"
	case kindFolder:
		return "Folder"
	}
	return ""
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
	scanned state = iota
	pending
	inProgress
	divergent
)

func (s state) String() string {
	switch s {
	case scanned:
		return "Scanned"
	case pending:
		return "Pending"
	case inProgress:
		return "In Progress"
	case divergent:
		return "Divergent"
	}
	return "UNKNOWN FILE STATE"
}

type sortColumn int

const (
	sortByName sortColumn = iota
	sortByTime
	sortBySize
)
