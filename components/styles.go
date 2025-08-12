package components

import "github.com/gdamore/tcell/v2"

var ModalStyle = tcell.StyleDefault.Foreground(tcell.ColorRed).Reverse(true)

type Styler interface {
	Style() tcell.Style
}
