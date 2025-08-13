package ui

import (
	"github.com/gdamore/tcell/v2"
)

// Constants really
var DefStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

var ViewStyle = DefStyle
var ViewDimmedStyle = DefStyle.Foreground(tcell.ColorDarkGray)
var CurrentMatchStyle = DefStyle.Foreground(tcell.ColorYellow)

var ViewLineNumberStyle = DefStyle.Foreground(tcell.ColorOrange)
var ViewDimmedLineNumberStyle = ViewLineNumberStyle.Foreground(tcell.ColorBrown)
var ViewCurrentMatchLineNumberStyle = DefStyle.Foreground(tcell.ColorYellow)

var ViewOverflowStyle = ViewStyle.Reverse(true)
var DimmedViewOverflowStyle = ViewOverflowStyle.Foreground(tcell.ColorDimGray)

var StatusBarStyle = tcell.StyleDefault.Reverse((true)).Bold((true))
var StatusBarBusyStyle = StatusBarStyle.Foreground(tcell.ColorRed)

// var TextInputStyle = tcell.StyleDefault.Background(tcell.ColorDarkBlue)
// var ActiveTextInputStyle = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack).Bold(true)
// var CursorTextInputStyle = ActiveTextInputStyle.Reverse(true)
