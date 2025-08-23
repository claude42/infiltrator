package ui

import (
	"github.com/claude42/infiltrator/components"
	"github.com/gdamore/tcell/v2"
)

type YesNoBar struct {
	components.ComponentImpl

	prompt  string
	yesFunc func()
	noFunc  func()
}

func ShowYesNoBar(prompt string, yesFunc func(), noFunc func()) {

	yn := &YesNoBar{
		prompt:  prompt,
		yesFunc: yesFunc,
		noFunc:  noFunc,
	}
	components.Add(yn, 2)
	width, height := screen.Size()
	yn.Resize(0, height-1, width, 1)
	yn.Show()
	yn.Render(true)
}

func (yn *YesNoBar) Resize(x, y, width, heigt int) {
	yn.ComponentImpl.Resize(0, y, width, 1)
}

func (yn *YesNoBar) Height() int {
	return 1
}

func (yn *YesNoBar) Size() (int, int) {
	return yn.Width(), 1
}

func (yn *YesNoBar) Render(updateScreen bool) {
	if !yn.IsVisible() {
		return
	}

	_, y := yn.Position()

	x := components.RenderText(0, y, yn.prompt, DefStyle)

	components.DrawChars(x, y, yn.Width(), ' ', DefStyle)

	if updateScreen {
		screen.Show()
	}
}

func (yn *YesNoBar) HandleEvent(ev tcell.Event) bool {
	if !yn.IsActive() {
		return false
	}

	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'y':
				yn.closeBar(yn.yesFunc)
				return true
			case 'n':
				yn.closeBar(yn.noFunc)
				return true
			}
		case tcell.KeyEnter:
			yn.closeBar(yn.yesFunc)
			return true
		case tcell.KeyEscape:
			yn.closeBar(yn.noFunc)
			return true
		}
	}
	return false
}

func (yn *YesNoBar) closeBar(f func()) {
	yn.Hide()
	components.Remove(yn)
	if f != nil {
		f()
	}
	components.RenderAll(true)
}
