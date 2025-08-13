package components

import "github.com/gdamore/tcell/v2"

var TextInputStyle = tcell.StyleDefault.Foreground((tcell.ColorDimGray))
var ActiveTextInputStyle = tcell.StyleDefault
var CursorTextInputStyle = ActiveTextInputStyle.Reverse(true)

var ModalStyle = tcell.StyleDefault.Foreground(tcell.ColorRed).Reverse(true)

type Styler interface {
	Style() tcell.Style
}
