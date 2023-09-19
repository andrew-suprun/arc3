package engine

import (
	r "arc/renderer"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

type (
	model struct {
		roots       []string
		archives    map[string]*archive
		filesByHash map[string][]*meta
		fsEvents    io.ReadCloser
		fsCommands  io.WriteCloser
		uiEvents    io.ReadCloser
		uiCommands  io.WriteCloser

		curRoot string
		curPath []string

		screenSize    r.Size
		fileTreeLines int

		quit                bool
		makeSelectedVisible bool
		requestFrame        bool
	}

	archive struct {
		root       string
		idx        int
		rootFolder *meta
		curFolder  *meta
		state      archiveState
		updated    bool
	}

	archiveState int

	state int

	kind int

	meta struct {
		kind     kind
		root     string
		name     string
		parent   *meta
		size     int
		modTime  time.Time
		state    state
		progress int
		hash     string
		counts   []int
		children map[string]*meta
	}
)

func (m *model) curArchive() *archive {
	return m.archives[m.curRoot]
}

func (m *meta) String() string {
	switch m.kind {
	case kindRegular:
		return fmt.Sprintf("file{ root=%q, folder=%q, name=%q, size=%d, mod-time=%s, state=%s, progress=%d, hash=%q, counts=%v }",
			m.root, filepath.Join(m.parent.path()...), m.name, m.size, m.modTime.Format(time.RFC3339), m.state, m.progress, m.hash, m.counts)
	case kindFolder:
		return fmt.Sprintf("file{ root=%q, folder=%q, name=%q, size=%d, mod-time=%s, state=%s, progress=%d, children=%d }",
			m.root, filepath.Join(m.parent.path()...), m.name, m.size, m.modTime.Format(time.RFC3339), m.state, m.progress, len(m.children))
	}
	panic("Invalid kind")
}

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

func counts(counts []int) string {
	if counts == nil {
		return ""
	}

	buf := &strings.Builder{}
	for _, count := range counts {
		fmt.Fprintf(buf, "%c", countRune(count))
	}
	return buf.String()
}

func countRune(count int) rune {
	if count == 0 {
		return '-'
	}
	if count > 9 {
		return '*'
	}
	return '0' + rune(count)
}
