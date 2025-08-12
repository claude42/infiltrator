package components

import "github.com/gdamore/tcell/v2"

var Screen tcell.Screen

func Beep() {
	Screen.Beep()
}
