package ui

import (
	"fmt"
	"math"
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
	styleFolderHeader   = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.ColorWhiteSmoke).Bold(true)
)

func (app *app) render() {
	if app.folderUpdateInProgress {
		return
	}
	folder := app.curFolder()
	lines := app.screenSize.height - 4
	entries := len(app.entries.entries)
	if folder.offsetIdx >= entries-lines {
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

	b := &builder{width: app.screenSize.width, height: app.screenSize.height, screen: app.screen, sync: app.sync}
	if app.screenSize.width < 80 || app.screenSize.height < 24 {
		b.space(app.screenSize.width, app.screenSize.height, styleScreenTooSmall)
		b.pos(app.screenSize.width/2-6, app.screenSize.height/2)
		b.layout(c{flex: 1})
		b.text("Too Small...", styleScreenTooSmall)
		b.show()
		return
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
	app.selectFolderTargets = app.selectFolderTargets[:0]
	layout := make([]c, 2*len(app.path)+2)
	layout[0] = c{size: 5}
	path := parsePath(app.path)
	for i, name := range path {
		nRunes := len([]rune(name))
		layout[2*i+1] = c{size: 3}
		layout[2*i+2] = c{size: nRunes}
	}
	layout[len(layout)-1] = c{flex: 1}
	b.layout(layout...)
	app.selectFolderTargets = append(app.selectFolderTargets, target{
		param:    "",
		position: position{x: 0, y: 1},
		size:     size{width: b.sizes[0], height: 1},
	})
	x := b.sizes[0] + b.sizes[1]
	for i := 2; i < len(b.sizes); i += 2 {
		app.selectFolderTargets = append(app.selectFolderTargets, target{
			param:    filepath.Join(path[:i/2]...),
			position: position{x: 1, y: 1},
			size:     size{width: b.sizes[i], height: 1},
		})
		x += b.sizes[i] + b.sizes[i+1]
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
	b.layout(c{size: 10}, c{size: 3}, c{size: 20, flex: 1}, c{size: 22}, c{size: 20}, c{size: 1})
	b.text(" State", styleFolderHeader)
	b.text("", styleFolderHeader)
	b.text("  Document"+folder.sortIndicator(sortByName), styleFolderHeader)
	b.text("  Date Modified"+folder.sortIndicator(sortByTime), styleFolderHeader)
	b.text(fmt.Sprintf("%20s", "Size"+folder.sortIndicator(sortBySize)), styleFolderHeader)
	b.text(" ", styleFolderHeader)
	lines := app.screenSize.height - 4

	for i := range app.entries.entries[folder.offsetIdx:] {
		entry := &app.entries.entries[folder.offsetIdx+i]
		if i >= lines {
			break
		}

		style := fileStyle(entry).Reverse(folder.selectedIdx == folder.offsetIdx+i)
		b.fileRow(entry, style)
	}
	rows := len(app.entries.entries) - folder.offsetIdx
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
	return tcell.StyleDefault.Foreground(tcell.PaletteColor(fg)).Background(17)
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

func progressBar(value float64, width int) string {
	if value < 0 || value > 1 {
		panic(fmt.Sprintf("Invalid progressBar value: %v", value))
	}
	if width < 1 {
		return ""
	}

	runes := make([]rune, width)
	progress := int(math.Round(float64(width*8) * float64(value)))
	idx := 0
	for ; idx < progress/8; idx++ {
		runes[idx] = '█'
	}
	if progress%8 > 0 {
		runes[idx] = []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}[progress%8]
		idx++
	}
	for ; idx < int(width); idx++ {
		runes[idx] = ' '
	}
	return string(runes)
}
