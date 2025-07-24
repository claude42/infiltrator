package ui

import (
	"github.com/gdamore/tcell/v2"
)

// Constants really
var DefStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
var LineNumberStyle = tcell.StyleDefault.Background(tcell.ColorOrange).Foreground(tcell.ColorBlack)
var OverflowStyle = tcell.StyleDefault.Reverse(true)
var TextInputStyle = tcell.StyleDefault.Background(tcell.ColorDarkBlue)
var ActiveTextInputStyle = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack).Bold(true)
var CursorTextInputStyle = ActiveTextInputStyle.Reverse(true)
