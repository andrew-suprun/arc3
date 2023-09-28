package ui

type folder struct {
	selectedIdx   int
	offsetIdx     int
	sortColumn    sortColumn
	sortAscending []bool
}

type sortColumn int

const (
	sortByName sortColumn = iota
	sortByTime
	sortBySize
)

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
	case sortByTime:
		app.entries.sortByTime()
	case sortBySize:
		app.entries.sortBySize()
	}
	if !folder.sortAscending[folder.sortColumn] {
		app.entries.reverse()
	}
}

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
