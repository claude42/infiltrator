package ui

import (
	//"log"

	"log"

	"github.com/claude42/infiltrator/util"
	// "github.com/claude42/infiltrator/model"

	"github.com/gdamore/tcell/v2"
)

const inputElementMargin = 3

type Input interface {
	SetContent(content string)
	Watch(eh tcell.EventHandler)
	SetColorIndex(colorIndex uint8)

	Component
}

type InputImpl struct {
	x, y, width int
	cursor      int
	start       int
	content     []rune

	inputCorrect bool
	colorIndex   uint8

	updateWatchers func()

	ComponentImpl
	util.ObservableImpl
}

func NewInputImpl() *InputImpl {
	i := &InputImpl{}
	i.inputCorrect = true
	i.updateWatchers = i.defaultUpdateWatchers

	return i
}

func (i *InputImpl) Resize(x, y, width, height int) {
	// height gets ignored
	i.x = x
	i.y = y
	i.width = width
	i.checkBoundaries()

}

func (i *InputImpl) SetContent(content string) {
	i.content = []rune(content)
	i.cursor = len(content)

	i.Render(true)
	i.updateWatchers()
}

func (i *InputImpl) Render(updateScreen bool) {
	style := i.determineStyle()

	x := renderRunes(i.x, i.y, i.width, i.content[i.start:], style)

	drawChars(x, i.y, i.x-x+i.width, '‾', style)

	if i.IsActive() {
		changeStyle(i.x+i.cursor-i.start, i.y, CursorTextInputStyle)
	}

	if updateScreen {
		screen.Show()
	}
}

func (i *InputImpl) SetColorIndex(colorIndex uint8) {
	i.colorIndex = colorIndex
}

func (i *InputImpl) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			i.insertRune(ev.Rune())
			return true
		case tcell.KeyLeft:
			if ev.Modifiers() == 0 {
				i.setCursor(i.cursor - 1)
				return true
			}
		case tcell.KeyRight:
			if ev.Modifiers() == 0 {
				i.setCursor(i.cursor + 1)
				return true
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			i.backspaceRune()
		case tcell.KeyDelete:
			i.deleteRune()
		case tcell.KeyCtrlU:
			i.deleteAll()
		default:
			//log.Printf("%v, %s", ev.Key(), tcell.KeyNames[ev.Key()])
		}
	}

	return false
}

func (i *InputImpl) insertRune(r rune) {
	var err error
	i.content, err = util.InsertRune(i.content, r, i.cursor)
	if err != nil {
		log.Panic("Input field out of bounds?!")
	}
	i.cursor++
	i.checkBoundaries()
	i.Render(true)
	i.updateWatchers()
}

func (i *InputImpl) setCursor(newCursor int) {
	if newCursor < 0 {
		i.cursor = 0
		screen.Beep()
	} else if newCursor > len(i.content) {
		i.cursor = len(i.content)
		screen.Beep()
	} else {
		i.cursor = newCursor
	}
	i.checkBoundaries()
	i.Render(true)
}

func (i *InputImpl) backspaceRune() {
	if i.cursor == 0 {
		screen.Beep()
		return
	}

	i.content = append(i.content[:i.cursor-1], i.content[i.cursor:]...)
	i.cursor--
	i.checkBoundaries()
	i.Render(true)
	i.updateWatchers()
}

func (i *InputImpl) deleteRune() {
	if i.cursor == len(i.content) {
		screen.Beep()
		return
	}

	i.content = append(i.content[:i.cursor], i.content[i.cursor+1:]...)
	i.checkBoundaries()
	i.Render(true)
	i.updateWatchers()
}

func (i *InputImpl) deleteAll() {
	i.content = i.content[:0]
	i.cursor = 0
	i.Render(true)
	i.updateWatchers()
}

func (i *InputImpl) checkBoundaries() {
	pos := i.cursor - i.start

	if pos >= i.width-inputElementMargin {
		i.start += pos - (i.width - inputElementMargin)
	} else if pos <= inputElementMargin-1 {
		i.start -= inputElementMargin - 1 - pos
		i.start = util.IntMax(i.start, 0)
	}

	// if i.cursor+i.start >= i.width-2 {
	// 	i.start = i.cursor - (i.width - 2)
	// } else if i.cursor <= 2+i.start {
	// 	i.start = i.cursor - 2
	// }
	// i.start, _ = util.InBetween(i.start, 0, util.IntMax(0, len(i.content)-i.width))
}

func (i *InputImpl) defaultUpdateWatchers() {
	ev := NewEventText(string(i.content))
	// consumed := i.PostEvent(ev)
	i.PostEvent(ev)

	// in case new inputCorrect state is different from previous
	// if consumed != i.inputCorrect {
	// 	i.inputCorrect = consumed
	// 	if !i.inputCorrect {
	// 		screen.Beep()
	// 	}
	// 	i.Render(true)
	// }
}

func (i *InputImpl) determineStyle() tcell.Style {
	var style tcell.Style

	if i.IsActive() {
		style = tcell.StyleDefault.Foreground((FilterColors[i.colorIndex][0]))
	} else {
		style = tcell.StyleDefault.Foreground((FilterColors[i.colorIndex][1]))
	}

	if !i.inputCorrect {
		style = style.Italic(true)
	}

	return style
}
