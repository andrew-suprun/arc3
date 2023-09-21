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
	styleFile           = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.ColorWhiteSmoke).Bold(true)
)

func (v *view) render(screen tcell.Screen) {
	folder := v.curFolder()
	lines := v.screenSize.height - 4
	entries := len(v.entries.entries)
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
	if v.makeSelectedVisible {
		if folder.offsetIdx <= folder.selectedIdx-lines {
			folder.offsetIdx = folder.selectedIdx + 1 - lines
		}
		if folder.offsetIdx > folder.selectedIdx {
			folder.offsetIdx = folder.selectedIdx
		}
		v.makeSelectedVisible = false
	}

	b := &builder{width: v.screenSize.width, height: v.screenSize.height, screen: screen, sync: v.sync}
	if v.screenSize.width < 80 || v.screenSize.height < 24 {
		b.space(v.screenSize.width, v.screenSize.height, styleScreenTooSmall)
		b.pos(v.screenSize.width/2-6, v.screenSize.height/2)
		b.layout(c{flex: 1})
		b.text("Too Small...", styleScreenTooSmall)
		b.show()
		return
	}
	v.showTitle(b)
	v.breadcrumbs(b)
	v.folderView(b)
	v.statusLine(b)
	b.show()
}

func (v *view) showTitle(b *builder) {
	b.layout(c{size: 9}, c{flex: 1})
	b.text(" Archive ", styleAppName)
	b.text(v.root, styleArchive)
}

func (v *view) breadcrumbs(b *builder) {
	b.newLine()
	v.selectFolderTargets = v.selectFolderTargets[:0]
	layout := make([]c, 2*len(v.path)+2)
	layout[0] = c{size: 5}
	path := parsePath(v.path)
	for i, name := range path {
		nRunes := len([]rune(name))
		layout[2*i+1] = c{size: 3}
		layout[2*i+2] = c{size: nRunes}
	}
	layout[len(layout)-1] = c{flex: 1}
	b.layout(layout...)
	v.selectFolderTargets = append(v.selectFolderTargets, target{
		param:    "",
		position: position{x: 0, y: 1},
		size:     size{width: b.sizes[0], height: 1},
	})
	x := b.sizes[0] + b.sizes[1]
	for i := 2; i < len(b.sizes); i += 2 {
		v.selectFolderTargets = append(v.selectFolderTargets, target{
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

func (v *view) folderView(b *builder) {
	b.newLine()
	folder := v.curFolder()
	b.layout(c{size: 10}, c{size: 3}, c{size: 20, flex: 1}, c{size: 22}, c{size: 20}, c{size: 1})
	b.text(" State", styleFolderHeader)
	b.text("", styleFolderHeader)
	b.text("  Document"+folder.sortIndicator(sortByName), styleFolderHeader)
	b.text("  Date Modified"+folder.sortIndicator(sortByTime), styleFolderHeader)
	b.text(fmt.Sprintf("%20s", "Size"+folder.sortIndicator(sortBySize)), styleFolderHeader)
	b.text(" ", styleFolderHeader)
	lines := v.screenSize.height - 4

	for i := range v.entries.entries[folder.offsetIdx:] {
		entry := &v.entries.entries[folder.offsetIdx+i]
		if i >= lines {
			break
		}

		style := v.fileStyle(entry).Reverse(folder.selectedIdx == folder.offsetIdx+i)
		b.fileRow(entry, style)
	}
	rows := len(v.entries.entries) - folder.offsetIdx
	if rows < lines {
		b.newLine()
		b.space(v.screenSize.width, lines-rows, styleDefault)
	}
}

func (v *view) fileStyle(file *entry) tcell.Style {
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

func (v *view) statusLine(b *builder) {
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
