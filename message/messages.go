package message

import (
	"fmt"
	"time"
)

type FolderScanned struct {
	Root string
	Path string
}

type FileScanned struct {
	Root    string
	Path    string
	Size    int
	ModTime time.Time
}

type ArchiveScanned struct {
	Root string
}

type ArchiveHashed struct {
	Root string
}

type FileHashed struct {
	Root string
	Path string
	Hash string
}

func (f FileHashed) String() string {
	return fmt.Sprintf("FileHashed: path: %q, hash: %q", f.Path, f.Hash)
}

type FileDeleted struct {
	Root string
	Path string
}

func (d FileDeleted) String() string {
	return fmt.Sprintf("FileDeleted: root: %q, name: %q", d.Root, d.Path)
}

type FileMoved struct {
	Root     string
	FromPath string
	ToPath   string
}

func (r FileMoved) String() string {
	return fmt.Sprintf("RenameFile: root: %q, from: %q, to: %q", r.Root, r.FromPath, r.ToPath)
}

type FileCopied struct {
	Root    string
	Path    string
	ToRoots []string
}

func (c FileCopied) String() string {
	return fmt.Sprintf("CopyFile: from: %q, to roots: %q", c.Path, c.ToRoots)
}

type ProgressState int // TODO move elsewhere

const (
	ProgressInitial ProgressState = iota
	ProgressScanned
	ProgressHashed
)

type HashingProgress struct {
	Root   string
	Path   string
	Hashed int
}

type CopyingProgress struct {
	Root   string
	Path   string
	Copied int
}

type Error struct {
	Path  string
	Error error
}

type ScreenSize struct {
	Width, Height int
}

type SelectArchive struct {
	Idx int
}

type Open struct{}

type Enter struct{}

type Exit struct{}

type RevealInFinder struct{}

type SelectFirst struct{}

type SelectLast struct{}

type MoveSelection struct{ Lines int }

type ResolveOne struct{}

type ResolveAll struct{}

type KeepAll struct{}

type Tab struct{}

type Delete struct{}

type Scroll struct {
	Command any
	Lines   int
}

func (s Scroll) String() string {
	return fmt.Sprintf("Scroll(%#v)", s.Lines)
}

type MouseTarget struct{ Command any }

func (t MouseTarget) String() string {
	return fmt.Sprintf("MouseTarget(%q)", t.Command)
}

type PgUp struct{}

type PgDn struct{}

type DebugPrintState struct{}

type DebugPrintRootWidget struct{}

type Quit struct{}
