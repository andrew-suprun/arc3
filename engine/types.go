package engine

import (
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

		quit bool
	}

	archive struct {
		root       string
		idx        int
		rootFolder *meta
		curFolder  *meta
		state      archiveState
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

func (m *meta) addChild(file *meta) {
	if m.children == nil {
		m.children = map[string]*meta{}
	}
	m.children[file.name] = file
}

func (m *meta) updateState() {
	if m == nil {
		return
	}
	m.size = 0
	m.progress = 0
	m.state = resolved
	m.modTime = nilTime
	for _, child := range m.children {
		m.progress += child.progress
		m.size += child.size
		if m.modTime.Before(child.modTime) {
			m.modTime = child.modTime
		}
		m.state = max(m.state, child.state)
	}
	m.parent.updateState()
}

var nilTime time.Time

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
		return "regular"
	case kindFolder:
		return "folder"
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
	inProgress
	pending
	divergent
)

func (s state) String() string {
	switch s {
	case resolved:
		return "resolved"
	case scanned:
		return "scanned"
	case inProgress:
		return "in-progress"
	case pending:
		return "pending"
	case divergent:
		return "divergent"
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
