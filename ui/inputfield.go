package ui

import (
	"log"

	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type InputField struct {
	x, y, width int
	cursor      int
	content     []rune
	receiver    model.UpdatedTextReceiver

	ComponentImpl
}

func NewInputField() *InputField {
	i := &InputField{}

	return i
}

func NewInputFieldWithReceiver(receiver model.UpdatedTextReceiver) *InputField {
	i := &InputField{}
	i.receiver = receiver

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
	i.updateReceiver()
}

func (i *InputField) Render(updateScreen bool) {
	var style tcell.Style
	if i.IsActive() {
		style = ActiveTextInputStyle
	} else {
		style = TextInputStyle
	}

	x := renderRunes(i.x, i.y, i.content, style)

	drawChars(x, i.y, i.width-len(i.content), ' ', style)

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
			i.setCursor(i.cursor - 1)
			return true
		case tcell.KeyRight:
			i.setCursor(i.cursor + 1)
			return true
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
	i.content = util.InsertRune(i.content, r, i.cursor)
	i.cursor++
	i.Render(true)
	i.updateReceiver()
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
	i.updateReceiver()
}

func (i *InputField) backspaceRune() {
	if i.cursor == 0 {
		screen.Beep()
		return
	}

	i.content = append(i.content[:i.cursor-1], i.content[i.cursor:]...)
	i.cursor--
	i.Render(true)
	i.updateReceiver()
}

func (i *InputField) deleteRune() {
	if i.cursor == len(i.content) {
		screen.Beep()
		return
	}

	i.content = append(i.content[:i.cursor], i.content[i.cursor+1:]...)
	i.Render(true)
	i.updateReceiver()
}

func (i *InputField) SetReceiver(receiver model.UpdatedTextReceiver) {
	i.receiver = receiver
}

func (i *InputField) updateReceiver() {
	log.Println(".")
	if i.receiver == nil {
		return
	}
	log.Println(".")

	i.receiver.UpdateText(string(i.content))
	log.Println(".")
}
