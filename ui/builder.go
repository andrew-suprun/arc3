package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type builder struct {
	x, y   int
	width  int
	height int
	field  int
	sizes  []int
	screen tcell.Screen
	sync   bool
}

type c struct {
	size, flex int
}

func (b *builder) pos(x, y int) {
	b.x = x
	b.y = y
}

func (b *builder) newLine() {
	b.x = 0
	b.y++
}

func (b *builder) layout(constraints ...c) []int {
	b.sizes = make([]int, len(constraints))
	b.field = 0
	totalSize, totalFlex := 0, 0
	for i, cons := range constraints {
		b.sizes[i] = cons.size
		totalSize += cons.size
		totalFlex += cons.flex
	}
	for totalSize > b.width {
		idx := 0
		maxSize := b.sizes[0]
		for i, size := range b.sizes {
			if maxSize < size {
				maxSize = size
				idx = i
			}
		}
		b.sizes[idx]--
		totalSize--
	}

	if totalFlex == 0 {
		return b.sizes
	}

	if totalSize < b.width {
		diff := b.width - totalSize
		remainders := make([]float64, len(constraints))
		for i, cons := range constraints {
			rate := float64(diff*cons.flex) / float64(totalFlex)
			remainders[i] = rate - math.Floor(rate)
			b.sizes[i] += int(rate)
		}
		totalSize := 0
		for _, size := range b.sizes {
			totalSize += size
		}
		for i := range b.sizes {
			if totalSize == b.width {
				break
			}
			if constraints[i].flex > 0 {
				b.sizes[i]++
				totalSize++
			}
		}
		for i := range b.sizes {
			if totalSize == b.width {
				break
			}
			if constraints[i].flex == 0 {
				b.sizes[i]++
				totalSize++
			}
		}
	}

	return b.sizes
}

func (b *builder) text(text string, style tcell.Style) {
	b.screen.SetCell(b.x, b.y, style, trim([]rune(text), b.sizes[b.field])...)
	b.field++
}

const modTimeFormat = "  " + time.RFC3339

func (b *builder) fileRow(file *entry, style tcell.Style) {
	b.newLine()
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

func (b *builder) state(file *entry, style tcell.Style) {
	switch file.state {
	case inProgress:
		value := float64(file.progress) / float64(file.size)
		b.progressBar(value, style)

	case scanned, pending:
		b.text(" ", style)

	case divergent:
		break

	default:
		return
	}
	b.text(file.counts, style)
}

func formatSize(size int) string {
	str := fmt.Sprintf("  %13d ", size)
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

func countRune(count int) rune {
	if count == 0 {
		return '-'
	}
	if count > 9 {
		return '*'
	}
	return '0' + rune(count)
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

func (b *builder) progressBar(value float64, style tcell.Style) {
	if value < 0 || value > 1 {
		panic(fmt.Sprintf("Invalid progressBar value: %v", value))
	}
	width := b.sizes[b.field]
	b.field++

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
	b.screen.SetCell(b.x, b.y, style, trim(runes, width)...)
}

func (b *builder) space(width, height int, style tcell.Style) {
	for line := b.y; line < b.y+height; line++ {
		for row := b.x; row < b.x+width; b.x++ {
			b.screen.SetCell(row, line, styleScreenTooSmall, ' ')
		}
	}
}

func trim(runes []rune, width int) []rune {
	if width < 1 {
		return nil
	}
	if len(runes) > int(width) {
		runes = append(runes[:width-1], '…')
	}
	diff := int(width) - len(runes)
	for diff > 0 {
		runes = append(runes, ' ')
		diff--
	}
	return runes
}

func (b *builder) show() {
	if b.sync {
		b.screen.Sync()
	} else {
		b.screen.Show()
	}
}
