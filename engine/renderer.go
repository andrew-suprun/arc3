package engine

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"
)

func (m *model) render() {
	if m.screenSize.Width < 80 || m.screenSize.Height < 24 {
		m.pos(0, 0)
		m.space(m.screenSize.Width, m.screenSize.Height, 9)
		m.pos(m.screenSize.Width/2-6, m.screenSize.Height/2)
		m.text("Too small...", 231, 9, bold)
		m.show()
		return
	}
	m.showTitle()
	m.breadcrumbs()
	m.folderView(m.curArchive.curFolder)
	m.statusLine()
	m.show()
}

func (m *model) showTitle() {
	sizes := calcSizes(m.screenSize.Width, c{size: 8}, c{size: 1}, c{flex: 1})
	m.pos(0, 0)
	m.text(text(" Archive", sizes[0])+text(" ", sizes[1]), 226, 0, bold+italic)
	m.text(text(m.curArchive.root, sizes[2]), 226, 0, bold)
}

func (m *model) breadcrumbs() {
	m.pos(0, 1)
	path := m.curArchive.curFolder.path()
	layout := make([]c, 2*len(path)+2)
	layout[0] = c{size: 5}
	size := 5
	for i, name := range path {
		nRunes := len([]rune(name))
		size += 3 + nRunes
		layout[2*i+1] = c{size: 3}
		layout[2*i+2] = c{size: nRunes}
	}
	layout[len(layout)-1] = c{flex: 1}
	sizes := calcSizes(m.screenSize.Width, layout...)
	m.text(text(" Root", sizes[0]), 250, 17, bold+italic)
	for i, name := range path {
		m.text(text(" / ", sizes[2*i+1]), 231, 17, bold)
		m.text(text(name, sizes[2*i+2]), 250, 17, bold+italic)
	}
	m.text(text("", sizes[len(layout)-1]), 0, 17, 0)
	x := 0
	for i := range path {
		folderPath := filepath.Join(path[:i]...)
		m.mouseTarget("select-folder", folderPath, x, 1, sizes[2*i]+x, 1)
	}
}

func (m *model) folderView(folder *folder) {
	entries := folder.entries()
	if entries == nil {
		return
	}
	sizes := calcSizes(m.screenSize.Width, c{size: 10}, c{size: 3}, c{size: 20, flex: 1}, c{size: 19}, c{size: 22})
	m.pos(0, 2)
	state := text(" State", sizes[0])
	kind := text("", sizes[1])
	document := text("  Document"+folder.sortIndicator(sortByName), sizes[2])
	date := text("  Date Modified"+folder.sortIndicator(sortByTime), sizes[3])
	size := text(fmt.Sprintf("%22s", "Size"+folder.sortIndicator(sortBySize)+" "), sizes[4])
	m.text(state+kind+document+date+size, 231, 8, bold)

	lines := m.screenSize.Height - 4
	m.fileTreeLines = lines
	if folder.offsetIdx >= len(entries)-lines {
		folder.offsetIdx = len(entries) + 1 - lines
	}
	if folder.offsetIdx < 0 {
		folder.offsetIdx = 0
	}
	if folder.selectedIdx >= len(entries) {
		folder.selectedIdx = len(entries) - 1
	}
	if folder.selectedIdx < 0 {
		folder.selectedIdx = 0
	}
	if m.makeSelectedVisible {
		if folder.offsetIdx <= folder.selectedIdx-lines {
			folder.offsetIdx = folder.selectedIdx + 1 - lines
		}
		if folder.offsetIdx > folder.selectedIdx {
			folder.offsetIdx = folder.selectedIdx
		}
		m.makeSelectedVisible = false
	}
	if folder.selectedName == "" {
		if len(entries) > 0 {
			entries[folder.selectedIdx].selected = true
			folder.selectedName = entries[folder.selectedIdx].name
		}
	} else {
		for i := range entries {
			if entries[i].name == folder.selectedName {
				folder.selectedIdx = i
				break
			}
		}
	}
	for i, entry := range entries[folder.offsetIdx:] {
		if i >= lines {
			break
		}

		flags := byte(0)
		if entry.selected {
			flags = reverse
		}
		row := m.fileRow(entry, sizes)
		m.pos(0, i+3)
		m.text(row, entry.statusFgColor(), 17, flags)
	}
	rows := len(entries) - folder.offsetIdx
	if rows < lines {
		m.pos(0, 3+rows)
		m.space(m.screenSize.Width, lines-rows, 17)
	}
}

func (m *model) fileRow(entry *entry, sizes []int) string {
	buf := &strings.Builder{}
	buf.WriteString(entry.stateText(sizes[0]))

	switch entry.kind {
	case kindRegular:
		buf.WriteString(text("", sizes[1]))
	case kindFolder:
		buf.WriteString(text(" ▶ ", sizes[1]))
	}

	buf.WriteString(text("  "+entry.name, sizes[2]))
	buf.WriteString(text("  "+entry.modTime.Format(time.DateTime), sizes[3]))
	buf.WriteString(text("  "+formatSize(entry.size), sizes[4]))

	return buf.String()
}

func (m *entry) stateText(size int) string {
	switch m.state {
	case hashing, copying:
		value := float64(m.progress) / float64(m.size)
		return " " + progressBar(value, size)
	case pending:
		return text(" Pending", size)
	case divergent:
		break
	default:
		return text("", size)
	}

	buf := &strings.Builder{}
	for _, count := range m.counts {
		fmt.Fprintf(buf, "%c", countRune(count))
	}
	return text(buf.String(), size)
}

func (m *model) statusLine() {
	m.pos(0, m.screenSize.Height-1)
	m.text(text(" Status will be here...", m.screenSize.Width), 231, 0, bold+italic)
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

func formatSize(size int) string {
	str := fmt.Sprintf("%13d ", size)
	slice := []string{str[:1], str[1:4], str[4:7], str[7:10]}
	b := strings.Builder{}
	for _, s := range slice {
		b.WriteString(s)
		if s == " " || s == "   " {
			b.WriteString(" ")
		} else {
			b.WriteString(",")
		}
	}
	b.WriteString(str[10:])
	return b.String()
}

func (e *entry) statusFgColor() (color byte) {
	switch e.state {
	case resolved, hashed:
		return 195
	case scanned, hashing:
		return 248
	case pending, copying:
		return 214
	case divergent:
		return 196
	}
	return 231
}

func (m *model) pos(x, y int) {
	m.sendToUi("pos", "x", x, "y", y)
}

func (m *model) text(text string, fg, bg, flags byte) {
	m.sendToUi("text", "text", text, "fg", fg, "bg", bg, "flags", flags)
}

func (m *model) space(width, height int, bg byte) {
	m.sendToUi("space", "width", width, "height", height, "bg", bg)
}

func (m *model) mouseTarget(command, path string, x, y, width, height int) {
	m.sendToUi("mouse-target", "command", command, "path", path, "x", x, "y", y, "width", width, "height", height)
}

func (m *model) scrollArea(command, path string, x1, y1, x2, y2 int) {
	m.sendToUi("scroll-area", "command", command, "path", path, "x1", x1, "y1", y1, "x2", x2, "y2", y2)
}

func (m *model) show() {
	m.sendToUi("show")
}

func (f *folder) sortIndicator(column sortColumn) string {
	if column == f.sortColumn {
		if f.sortAscending[column] {
			return " ▲"
		}
		return " ▼"
	}
	return ""
}

type c struct {
	size, flex int
}

var (
	bold    byte = 1
	italic  byte = 2
	reverse byte = 4
)

func calcSizes(targetSize int, constraints ...c) []int {
	result := make([]int, len(constraints))
	totalSize, totalFlex := 0, 0
	for i, cons := range constraints {
		result[i] = cons.size
		totalSize += cons.size
		totalFlex += cons.flex
	}
	for totalSize > targetSize {
		idx := 0
		maxSize := result[0]
		for i, size := range result {
			if maxSize < size {
				maxSize = size
				idx = i
			}
		}
		result[idx]--
		totalSize--
	}

	if totalFlex == 0 {
		return result
	}

	if totalSize < targetSize {
		diff := targetSize - totalSize
		remainders := make([]float64, len(constraints))
		for i, cons := range constraints {
			rate := float64(diff*cons.flex) / float64(totalFlex)
			remainders[i] = rate - math.Floor(rate)
			result[i] += int(rate)
		}
		totalSize := 0
		for _, size := range result {
			totalSize += size
		}
		for i := range result {
			if totalSize == targetSize {
				break
			}
			if constraints[i].flex > 0 {
				result[i]++
				totalSize++
			}
		}
		for i := range result {
			if totalSize == targetSize {
				break
			}
			if constraints[i].flex == 0 {
				result[i]++
				totalSize++
			}
		}
	}

	return result
}

func text(text string, width int) string {
	if width < 1 {
		return ""
	}
	runes := []rune(text)
	if len(runes) > int(width) {
		runes = append(runes[:width-1], '…')
	}
	diff := int(width) - len(runes)
	for diff > 0 {
		runes = append(runes, ' ')
		diff--
	}
	return string(runes)
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
