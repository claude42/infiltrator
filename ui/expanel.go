package ui

import (
	"github.com/claude42/infiltrator/components"

	"github.com/gdamore/tcell/v2"
)

type ExPanel struct {
	*components.PanelImpl

	input *components.InputImpl
	// mode          *ColoredSelect
	// caseSensitive *ColoredSelect
}

func NewExPanel() *ExPanel {
	e := &ExPanel{
		PanelImpl: components.NewPanelImpl("none"),
		input:     components.NewInputImpl(),
	}
	return e
}

func (e *ExPanel) Resize(x, y, width, height int) {
	e.PanelImpl.Resize(x, y, width, height)

	e.input.Resize(x+2, y, width-2, 1)
}

func (e *ExPanel) Render(updateScreen bool) {
	if !e.IsVisible() {
		return
	}

	style := e.PanelImpl.CurrentStyler.Style()

	x, y := e.Position()

	components.RenderText(x, y, ":", style)

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

func (e *ExPanel) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			e.SetActive(false)
		}
	}

	if e.input != nil && e.input.HandleEvent(ev) {
		return true
	}

	return false
}

func (e *ExPanel) SetActive(active bool) {
	e.PanelImpl.SetActive(active)

	e.input.SetActive(active)
}

// func (t *StringFilterPanel) WatchInput(eh tcell.EventHandler) {
// 	if t.input == nil {
// 		log.Panicln("StringFilterPanel.WatchInput() called without input field!")
// 		return
// 	}
// 	t.input.Watch(eh)
// }
