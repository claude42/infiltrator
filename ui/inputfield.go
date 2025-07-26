package ui

import (
	"log"

	// "github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type InputField struct {
	x, y, width int
	cursor      int
	content     []rune

	inputCorrect bool
	colorIndex   uint8

	ComponentImpl
	util.ObservableImpl
}

func NewInputField() *InputField {
	i := &InputField{}
	i.inputCorrect = true

	return i
}

func NewInputFieldWithEventhandler(eh tcell.EventHandler) *InputField {
	i := &InputField{}
	i.Watch(eh)
	i.inputCorrect = true

	return i
}

func (i *InputField) Resize(x, y, width, height int) {
	// height gets ignored
	i.x = x
	i.y = y
	i.width = width

}

func (i *InputField) SetContent(content string) {
	i.content = []rune(content)
	i.cursor = len(content)

	i.Render(true)
	i.updateWatchers()
}

func (i *InputField) Render(updateScreen bool) {
	style := i.determineStyle()

	x := renderRunes(i.x, i.y, i.content, style)

	drawChars(x, i.y, i.width-len(i.content), '‾', style)

	if i.IsActive() {
		changeStyle(i.x+i.cursor, i.y, CursorTextInputStyle)
	}

	if updateScreen {
		screen.Show()
	}
}

func (i *InputField) HandleEvent(ev tcell.Event) bool {
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
		default:
			//log.Printf("%v, %s", ev.Key(), tcell.KeyNames[ev.Key()])
		}
	}

	return false
}

func (i *InputField) insertRune(r rune) {
	var err error
	i.content, err = util.InsertRune(i.content, r, i.cursor)
	if err != nil {
		log.Panic("Input field out of bounds?!")
	}
	i.cursor++
	i.Render(true)
	i.updateWatchers()
}

func (i *InputField) setCursor(newCursor int) {
	if newCursor < 0 {
		i.cursor = 0
		screen.Beep()
	} else if newCursor > len(i.content) {
		i.cursor = len(i.content)
		screen.Beep()
	} else {
		i.cursor = newCursor
	}
	i.Render(true)
	i.updateWatchers()
}

func (i *InputField) backspaceRune() {
	if i.cursor == 0 {
		screen.Beep()
		return
	}

	i.content = append(i.content[:i.cursor-1], i.content[i.cursor:]...)
	i.cursor--
	i.Render(true)
	i.updateWatchers()
}

func (i *InputField) deleteRune() {
	if i.cursor == len(i.content) {
		screen.Beep()
		return
	}

	i.content = append(i.content[:i.cursor], i.content[i.cursor+1:]...)
	i.Render(true)
	i.updateWatchers()
}

func (i *InputField) updateWatchers() {
	ev := util.NewEventText(string(i.content))
	consumed := i.PostEvent(ev)

	// in case new inputCorrect state is different from previous
	if consumed != i.inputCorrect {
		i.inputCorrect = consumed
		screen.Beep()
		i.Render(true)
	}
}

func (i *InputField) SetColorIndex(colorIndex uint8) {
	i.colorIndex = colorIndex
}

func (i *InputField) determineStyle() tcell.Style {
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
