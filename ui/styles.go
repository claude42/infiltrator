package ui

import (
	"github.com/gdamore/tcell/v2"
)

// Constants really
var DefStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

var ViewStyle = DefStyle
var ViewDimmedStyle = DefStyle.Foreground(tcell.ColorGray)

var ViewLineNumberStyle = DefStyle.Foreground(tcell.ColorOrange)
var ViewDimmedLineNumberStyle = ViewLineNumberStyle.Foreground(tcell.ColorBrown)

var ViewOverflowStyle = ViewStyle.Reverse(true)
var DimmedViewOverflowStyle = ViewOverflowStyle.Foreground(tcell.ColorDimGray)

var TextInputStyle = tcell.StyleDefault.Background(tcell.ColorDarkBlue)
var ActiveTextInputStyle = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack).Bold(true)
var CursorTextInputStyle = ActiveTextInputStyle.Reverse(true)
