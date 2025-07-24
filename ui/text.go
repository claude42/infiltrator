package ui

import (
	//"log"

	"github.com/gdamore/tcell/v2"
)

func renderRunes(x int, y int, runes []rune, style tcell.Style) int {
	maxWidth, _ := screen.Size()
	for _, r := range runes {
		if x >= maxWidth {
			break
		}
		screen.SetContent(x, y, r, nil, style)
		x++
	}

	return x
}

func renderText(x int, y int, text string, style tcell.Style) int {
	return renderRunes(x, y, []rune(text), style)
}

func drawChars(x int, y int, width int, r rune, style tcell.Style) int {
	screenWidth, _ := screen.Size()
	i := 0
	for ; i < width && x+i < screenWidth; i++ {
		screen.SetContent(x+i, y, r, nil, style)
	}

	return x + i
}

func fillChars(x int, y int, r rune, style tcell.Style) {
	maxWidth, _ := screen.Size()

	for i := x; i <= maxWidth; i++ {
		screen.SetContent(i, y, r, nil, style)
	}
}

func changeStyle(x int, y int, style tcell.Style) {
	r, _, _, _ := screen.GetContent(x, y)
	screen.SetContent(x, y, r, nil, style)
}
