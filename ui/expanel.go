package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/components"

	"github.com/gdamore/tcell/v2"
)

type ExPanel struct {
	*components.PanelImpl

	prompt string

	input *components.InputImpl
	// mode          *ColoredSelect
	// caseSensitive *ColoredSelect
}

func NewExPanel() *ExPanel {
	e := &ExPanel{
		PanelImpl: components.NewPanelImpl("none"),
		input:     components.NewInputImpl(),
	}
	e.PanelImpl.Add(e.input)
	return e
}

func (e *ExPanel) Resize(x, y, width, height int) {
	e.PanelImpl.Resize(x, y, width, 1)

	// reload in case Resize() was called with zero values
	x, y = e.PanelImpl.Position()

	e.input.Resize(x+len(e.prompt)+2, y, e.PanelImpl.Width()-len(e.prompt)-2, 1)
}

func (e *ExPanel) Render(updateScreen bool) {
	if !e.IsVisible() {
		return
	}

	style := e.PanelImpl.CurrentStyler.Style()

	x, y := e.Position()

	fullPrompt := fmt.Sprint(e.prompt + ": ")

	components.RenderText(x, y, fullPrompt, style)

	if e.input != nil {
		e.input.Render((updateScreen))
	}

	if updateScreen {
		screen.Show()
	}
}

func (e *ExPanel) SetContent(content string) {
	// fail.IfNil(t.input, "ExPanel.SetContent() called without input field!")

	e.input.SetContent(content)
}

func (e *ExPanel) SetPrompt(prompt string) {
	e.prompt = prompt
	e.Resize(-1, -1, -1, -1)
	e.Render(true)
}

func (e *ExPanel) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			e.SetActive(false)
			return true
		}
	}

	if e.input != nil && e.input.HandleEvent(ev) {
		return true
	}

	return false
}

// func (t *StringFilterPanel) WatchInput(eh tcell.EventHandler) {
// 	if t.input == nil {
// 		log.Panicln("StringFilterPanel.WatchInput() called without input field!")
// 		return
// 	}
// 	t.input.Watch(eh)
// }
