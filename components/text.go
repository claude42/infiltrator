package components

import (
	"github.com/gdamore/tcell/v2"
)

func RenderRunes(x int, y int, width int, runes []rune, style tcell.Style) int {
	if len(runes) == 0 {
		return x
	}

	var i int
	var r rune
	for i, r = range runes {
		if i >= width {
			break
		}
		Screen.SetContent(x+i, y, r, nil, style)
	}

	return x + i + 1
}

func RenderText(x int, y int, text string, style tcell.Style) int {
	// TODO: should also integrate a width here at some point in time
	width, _ := Screen.Size()
	return RenderRunes(x, y, width, []rune(text), style)
}

func centerText(x int, width int, y int, text string, style tcell.Style) int {

	return RenderText(x+(width-len(text))/2, y, text, style)
}

func DrawChars(x int, y int, width int, r rune, style tcell.Style) int {
	screenWidth, _ := Screen.Size()
	i := 0
	for ; i < width && x+i < screenWidth; i++ {
		Screen.SetContent(x+i, y, r, nil, style)
	}

	return x + i
}

// func fillChars(x int, y int, r rune, style tcell.Style) {
// 	maxWidth, _ := screen.Size()

// 	for i := x; i <= maxWidth; i++ {
// 		screen.SetContent(i, y, r, nil, style)
// 	}
// }

func ChangeStyle(x int, y int, style tcell.Style) {
	r, _, _, _ := Screen.GetContent(x, y)
	Screen.SetContent(x, y, r, nil, style)
}
