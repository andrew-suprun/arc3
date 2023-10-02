package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	styleDefault        = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.Color17).Bold(true)
	styleScreenTooSmall = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.ColorRed).Bold(true)
	styleAppName        = tcell.StyleDefault.Foreground(tcell.Color226).Background(tcell.ColorBlack).Bold(true).Italic(true)
	styleArchive        = tcell.StyleDefault.Foreground(tcell.Color226).Background(tcell.ColorBlack).Bold(true)
	styleBreadcrumbs    = tcell.StyleDefault.Foreground(tcell.Color250).Background(tcell.Color17).Bold(true).Italic(true)
	styleFolderHeader   = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.ColorGray).Bold(true)
	styleProgressBar    = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.ColorLightGray)
)

func (app *app) render() {
	if app.folderUpdateInProgress {
		return
	}

	b := &builder{width: app.screenSize.width, height: app.screenSize.height, screen: app.screen, sync: app.sync}
	app.sync = false

	if app.screenSize.width < 80 || app.screenSize.height < 24 {
		b.space(app.screenSize.width, app.screenSize.height, styleScreenTooSmall)
		b.pos(app.screenSize.width/2-6, app.screenSize.height/2)
		b.layout(c{flex: 1})
		b.text("Too Small...", styleScreenTooSmall)
		b.show()
		return
	}

	folder := app.curFolder()
	lines := app.screenSize.height - 4
	entries := len(app.entries)
	if folder.offsetIdx >= entries-lines+1 {
		folder.offsetIdx = entries + 1 - lines
	}
	if folder.offsetIdx < 0 {
		folder.offsetIdx = 0
	}
	if folder.selectedIdx >= entries {
		folder.selectedIdx = entries - 1
	}
	if folder.selectedIdx < 0 {
		folder.selectedIdx = 0
	}
	if app.makeSelectedVisible {
		if folder.offsetIdx <= folder.selectedIdx-lines {
			folder.offsetIdx = folder.selectedIdx + 1 - lines
		}
		if folder.offsetIdx > folder.selectedIdx {
			folder.offsetIdx = folder.selectedIdx
		}
		app.makeSelectedVisible = false
	}

	app.showTitle(b)
	app.breadcrumbs(b)
	app.folderView(b)
	app.statusLine(b)
	b.show()
}

func (app *app) showTitle(b *builder) {
	b.layout(c{size: 9}, c{flex: 1})
	b.text(" Archive ", styleAppName)
	b.text(app.root, styleArchive)
}

func (app *app) breadcrumbs(b *builder) {
	b.newLine()
	app.folderTargets = app.folderTargets[:0]
	path := parsePath(app.curPath())
	layout := make([]c, 2*len(path)+2)
	layout[0] = c{size: 5}
	for i, name := range path {
		nRunes := len([]rune(name))
		layout[2*i+1] = c{size: 3}
		layout[2*i+2] = c{size: nRunes}
	}
	layout[len(layout)-1] = c{size: 1, flex: 1}
	b.layout(layout...)
	app.folderTargets = append(app.folderTargets, folderTarget{
		param:  "",
		offset: 0,
		width:  b.widths[0],
	})
	x := b.widths[0] + b.widths[1]
	for i := 2; i < len(b.widths); i += 2 {
		app.folderTargets = append(app.folderTargets, folderTarget{
			param:  filepath.Join(path[:i/2]...),
			offset: b.offsets[i],
			width:  b.widths[i],
		})
		x += b.widths[i] + b.widths[i+1]
	}
	b.text(" Root", styleBreadcrumbs)
	for _, name := range path {
		b.text(" / ", styleDefault)
		b.text(name, styleBreadcrumbs)
	}
	b.text("", styleBreadcrumbs)
}

func (app *app) folderView(b *builder) {
	b.newLine()
	folder := app.curFolder()
	b.layout(c{size: 1}, c{size: 10}, c{size: 3}, c{size: 20, flex: 1}, c{size: 22}, c{size: 19}, c{size: 1})
	app.sortTargets = make([]sortTarget, 3)
	app.sortTargets[0] = sortTarget{
		sortColumn: sortByName,
		offset:     b.offsets[3],
		width:      b.widths[3],
	}
	app.sortTargets[1] = sortTarget{
		sortColumn: sortByTime,
		offset:     b.offsets[4],
		width:      b.widths[4],
	}
	app.sortTargets[2] = sortTarget{
		sortColumn: sortBySize,
		offset:     b.offsets[5],
		width:      b.widths[5],
	}
	b.text(" ", styleFolderHeader)
	b.text("State", styleFolderHeader)
	b.text("", styleFolderHeader)
	b.text("Document"+folder.sortIndicator(sortByName), styleFolderHeader)
	b.text("  Date Modified"+folder.sortIndicator(sortByTime), styleFolderHeader)
	b.text(fmt.Sprintf("%19s", "Size"+folder.sortIndicator(sortBySize)), styleFolderHeader)
	b.text(" ", styleFolderHeader)
	lines := app.screenSize.height - 4

	for i := range app.entries[folder.offsetIdx:] {
		file := &app.entries[folder.offsetIdx+i]
		if i >= lines {
			break
		}

		style := fileStyle(file).Reverse(folder.selectedIdx == folder.offsetIdx+i)
		b.newLine()
		b.text(" ", style)
		b.state(file, style)
		switch file.kind {
		case kindRegular:
			b.text("   ", style)
		case kindFolder:
			b.text(" ▶ ", style)
		}
		b.text(file.name, style)
		b.text(file.modTime.Format(modTimeFormat), style)
		b.text(formatSize(file.size), style)
		b.text(" ", style)
	}
	rows := len(app.entries) - folder.offsetIdx
	if rows < lines {
		b.newLine()
		b.space(app.screenSize.width, lines-rows, styleDefault)
	}
}

func fileStyle(file *entry) tcell.Style {
	fg := 231
	switch file.state {
	case scanned:
		fg = 248
	case inProgress:
		fg = 195
	case pending:
		fg = 214
	case divergent:
		fg = 196
	}
	return tcell.StyleDefault.Foreground(tcell.PaletteColor(fg)).Background(tcell.PaletteColor(17))
}

func (app *app) statusLine(b *builder) {
	b.newLine()
	b.layout(c{flex: 1})
	b.text(" Status line will be here...", styleArchive)
}

func parsePath(strPath string) []string {
	path := strings.Split(string(strPath), "/")
	if path[0] != "" {
		return path
	}
	return nil
}
