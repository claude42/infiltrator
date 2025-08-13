package components

import (
	"sync"

	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

const inputElementMargin = 3

type Input interface {
	Component

	SetContent(content string)
	Watch(eh tcell.EventHandler)
	// SetColorIndex(colorIndex uint8)
}

type InputImpl struct {
	ComponentImpl
	util.ObservableImpl
	sync.Mutex

	x, y, width int
	cursor      int
	start       int
	content     []rune

	InputCorrect bool

	delay                 *util.Delay
	OldUpdateWatchersFunc func()
	UpdateWatchersFunc    func()

	OldStyler     Styler
	CurrentStyler Styler
}

func NewInputImpl() *InputImpl {
	i := &InputImpl{}
	i.InputCorrect = true
	i.delay = util.NewDelay(i.DefaultUpdateWatchers)
	i.UpdateWatchersFunc = i.DefaultUpdateWatchers

	i.StyleUsing(i)

	return i
}

func (i *InputImpl) SetUpdateWatchersFunc(newFunc func()) {
	if i.UpdateWatchersFunc != nil {
		i.OldUpdateWatchersFunc = i.UpdateWatchersFunc
	}
	i.UpdateWatchersFunc = newFunc
	i.delay.SetInvokeFunc(newFunc)
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
	i.cursor = len(i.content)

	i.Render(true)
	i.UpdateWatchersFunc()
}

func (i *InputImpl) Content() string {
	return string(i.content)
}

func (i *InputImpl) SetActive(active bool) {
	i.ComponentImpl.SetActive(active)

	i.Render(true)
}

func (i *InputImpl) Render(updateScreen bool) {
	if !i.visible {
		return
	}

	style := i.CurrentStyler.Style()

	x := RenderRunes(i.x, i.y, i.width, i.content[i.start:], style)

	DrawChars(x, i.y, i.x-x+i.width, '‾', style)

	if i.IsActive() {
		ChangeStyle(i.x+i.cursor-i.start, i.y, CursorTextInputStyle)
	}

	if updateScreen {
		Screen.Show()
	}
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
			i.deleteToTheLeft()
		case tcell.KeyCtrlK:
			i.deleteToTheRight()
		case tcell.KeyCtrlA:
			i.setCursor(0)
			return true
		case tcell.KeyCtrlE:
			i.setCursor(int(len(i.content)))
			return true
		default:
			//log.Printf("%v, %s", ev.Key(), tcell.KeyNames[ev.Key()])
		}
	}

	return false
}

func (i *InputImpl) insertRune(r rune) {
	var err error
	i.content, err = util.InsertRune(i.content, r, i.cursor)
	fail.OnError(err, "Input field out of bounds?!")

	i.cursor++
	i.checkBoundaries()
	i.Render(true)
	i.delay.Now()
}

func (i *InputImpl) setCursor(newCursor int) {
	if newCursor < 0 {
		i.cursor = 0
		Screen.Beep()
	} else if newCursor > len(i.content) {
		i.cursor = len(i.content)
		Screen.Beep()
	} else {
		i.cursor = newCursor
	}
	i.checkBoundaries()
	i.Render(true)
}

func (i *InputImpl) backspaceRune() {
	if i.cursor == 0 {
		Screen.Beep()
		return
	}

	i.content = append(i.content[:i.cursor-1], i.content[i.cursor:]...)
	i.cursor--
	i.checkBoundaries()
	i.Render(true)
	i.delay.Now()
}

func (i *InputImpl) deleteRune() {
	if i.cursor == len(i.content) {
		Screen.Beep()
		return
	}

	i.content = append(i.content[:i.cursor], i.content[i.cursor+1:]...)
	i.checkBoundaries()
	i.Render(true)
	i.delay.Now()
}

func (i *InputImpl) deleteToTheLeft() {
	if i.cursor >= len(i.content) {
		i.content = []rune("")
	} else {
		i.content = i.content[i.cursor:]
	}
	i.cursor = 0
	i.Render(true)
	i.delay.Now()
}

func (i *InputImpl) deleteToTheRight() {
	if i.cursor >= len(i.content) {
		return
	}

	i.content = i.content[:i.cursor]
	i.Render(true)
	i.delay.Now()
}

func (i *InputImpl) checkBoundaries() {
	pos := i.cursor - i.start

	if pos >= i.width-inputElementMargin {
		i.start += pos - (i.width - inputElementMargin)
	} else if pos <= inputElementMargin-1 {
		i.start -= inputElementMargin - 1 - pos
		i.start = max(i.start, 0)
	}

}

func (i *InputImpl) DefaultUpdateWatchers() {
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

func (i *InputImpl) Style() tcell.Style {
	var style tcell.Style

	if !i.InputCorrect {
		style = style.Italic(true)
	}

	return style
}

func (i *InputImpl) StyleUsing(styler Styler) {
	if i.CurrentStyler != nil {
		i.OldStyler = i.CurrentStyler
	}

	i.CurrentStyler = styler
}
