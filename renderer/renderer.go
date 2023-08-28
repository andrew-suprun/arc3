package renderer

import (
	"math"
)

type Position struct {
	X, Y int
}

type Size struct {
	Width, Height int
}

type Style struct {
	FG, BG byte
	Flags  Flags
}

type Flags byte

const (
	Bold    Flags = 1
	Italic  Flags = 2
	Reverse Flags = 4
)

type Renderer func(screen UI, pos Position, size Size)

type UI interface {
	Text(text string, pos Position, style Style)
	Space(pos Position, size Size, style Style)
	MouseTarget(command any, pos Position, size Size)
	Scroll(command any, pos Position, size Size)
	Render()
}

func Column(layout Layout, renderers ...Renderer) Renderer {
	if len(layout) != len(renderers) {
		panic("Column inconsistent sizes")
	}
	return func(screen UI, pos Position, size Size) {
		heights := layout.calcSizes(size.Height)
		y := pos.Y
		for i, height := range heights {
			renderers[i](screen, Position{X: pos.X, Y: y}, Size{Width: size.Width, Height: height})
			y += height
		}
	}
}

func Row(layout Layout, renderers ...Renderer) Renderer {
	if len(layout) != len(renderers) {
		panic("Row inconsistent sizes")
	}
	return func(screen UI, pos Position, size Size) {
		widths := layout.calcSizes(size.Width)
		x := pos.X
		for i, width := range widths {
			renderers[i](screen, Position{X: x, Y: pos.Y}, Size{Width: width, Height: size.Height})
			x += width
		}
	}
}

func Text(text string, style Style) Renderer {
	runes := []rune(text)
	return func(screen UI, pos Position, size Size) {
		if size.Width < 1 {
			return
		}
		if len(runes) > int(size.Width) {
			runes = append(runes[:size.Width-1], '…')
		}
		diff := int(size.Width) - len(runes)
		for diff > 0 {
			runes = append(runes, ' ')
			diff--
		}

		screen.Text(string(runes), pos, style)
	}
}

func ProgressBar(value float64, style Style) Renderer {
	return func(screen UI, pos Position, size Size) {
		if size.Width < 1 {
			return
		}

		runes := make([]rune, size.Width)
		progress := int(math.Round(float64(size.Width*8) * float64(value)))
		idx := 0
		for ; idx < progress/8; idx++ {
			runes[idx] = '█'
		}
		if progress%8 > 0 {
			runes[idx] = []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}[progress%8]
			idx++
		}
		for ; idx < int(size.Width); idx++ {
			runes[idx] = ' '
		}

		screen.Text(string(runes), pos, style)
	}
}

func Spacer(style Style) Renderer {
	return func(screen UI, pos Position, size Size) {
		screen.Space(pos, size, style)
	}
}

func MouseTarget(cmd any, child Renderer) Renderer {
	return func(screen UI, pos Position, size Size) {
		screen.MouseTarget(cmd, pos, size)
		child(screen, pos, size)
	}
}

func Scroll(cmd string, child func(height int) Renderer) Renderer {
	return func(screen UI, pos Position, size Size) {
		screen.Scroll(cmd, pos, size)
		child(size.Height)(screen, pos, size)
	}
}
