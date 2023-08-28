package engine

import (
	r "arc/renderer"
	"io"
	"time"
)

type (
	model struct {
		roots      []string
		archives   map[string]*archive
		fsEvents   io.ReadCloser
		fsCommands io.WriteCloser
		uiEvents   io.ReadCloser
		uiCommands io.WriteCloser

		screenSize    r.Size
		fileTreeLines int

		curArchive          *archive
		quit                bool
		makeSelectedVisible bool
		requestFrame        bool
	}

	archive struct {
		root       string
		idx        int
		rootFolder *folder
		curFolder  *folder
		state      archiveState
		updated    bool
	}

	archiveState int

	meta struct {
		root     string
		name     string
		parent   *folder
		size     int
		modTime  time.Time
		state    state
		progress int
	}

	file struct {
		meta
		hash   string
		counts []int
	}

	folder struct {
		meta
		children      map[string]*folder
		files         map[string]*file
		selectedName  string
		selectedIdx   int
		offsetIdx     int
		sortColumn    sortColumn
		sortAscending []bool
	}

	entry struct {
		kind     kind
		name     string
		size     int
		modTime  time.Time
		state    state
		progress int
		counts   []int
		selected bool
	}

	kind int

	state int

	builder struct {
		width  int
		height int
		out    io.Writer
	}

	pos struct {
		x, y int
	}

	style struct {
		fg, bg, flags byte
	}

	sortColumn int
)

const (
	kindRegular kind = iota
	kindFolder
)

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

const (
	sortByName sortColumn = iota
	sortByTime
	sortBySize
)

func (c sortColumn) String() string {
	switch c {
	case sortByName:
		return "SortByName"
	case sortByTime:
		return "SortByTime"
	case sortBySize:
		return "SortBySize"
	}
	return "Illegal Sort Solumn"
}

const (
	resolved state = iota
	scanned
	hashing
	hashed
	pending
	copying
	copied
	divergent
)

func (s state) String() string {
	switch s {
	case resolved:
		return "Resolved"
	case scanned:
		return "Scanned"
	case hashing:
		return "Hashing"
	case hashed:
		return "Hashed"
	case pending:
		return "Pending"
	case copying:
		return "Copying"
	case copied:
		return "Copied"
	case divergent:
		return "Divergent"
	}
	return "UNKNOWN FILE STATE"
}
