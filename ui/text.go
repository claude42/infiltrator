package ui

import (
	//"log"

	"github.com/gdamore/tcell/v2"
)

func renderRunes(x int, y int, width int, runes []rune, style tcell.Style) int {
	if len(runes) == 0 {
		return x
	}

	var i int
	var r rune
	for i, r = range runes {
		if i >= width {
			break
		}
		screen.SetContent(x+i, y, r, nil, style)
	}

	return x + i + 1
}

func renderText(x int, y int, text string, style tcell.Style) int {
	// TODO: should also integrate a width here at some point in time
	width, _ := screen.Size()
	return renderRunes(x, y, width, []rune(text), style)
}

func centerText(x int, width int, y int, text string, style tcell.Style) int {

	return renderText(x+(width-len(text))/2, y, text, style)
}

func drawChars(x int, y int, width int, r rune, style tcell.Style) int {
	screenWidth, _ := screen.Size()
	i := 0
	for ; i < width && x+i < screenWidth; i++ {
		screen.SetContent(x+i, y, r, nil, style)
	}

	return x + i
}

// func fillChars(x int, y int, r rune, style tcell.Style) {
// 	maxWidth, _ := screen.Size()

// 	for i := x; i <= maxWidth; i++ {
// 		screen.SetContent(i, y, r, nil, style)
// 	}
// }

func changeStyle(x int, y int, style tcell.Style) {
	r, _, _, _ := screen.GetContent(x, y)
	screen.SetContent(x, y, r, nil, style)
}
