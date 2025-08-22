package ui

import (
	"github.com/claude42/infiltrator/components"
	"github.com/gdamore/tcell/v2"
)

type QuestionBar struct {
	*components.PanelImpl

	prompt string

	input *components.InputImpl

	enterFunc func(value string)
}

func ShowQuestionBar(prompt string, defaultValue string,
	enterFunc func(value string)) {

	q := &QuestionBar{
		PanelImpl: components.NewPanelImpl("none"),
		input:     components.NewInputImpl(),
		prompt:    prompt,
		enterFunc: enterFunc,
	}
	q.Add(q.input)
	q.input.SetContent(defaultValue)
	components.Add(q, 2)
	width, height := screen.Size()
	q.Resize(0, height-1, width, 1)
	q.Show()
	q.Render(true)
}

func (q *QuestionBar) Resize(x, y, width, heigt int) {
	q.ComponentImpl.Resize(0, y, width, 1)

	q.input.Resize(len(q.prompt), y, width-(len(q.prompt)), 1)
}

func (q *QuestionBar) Height() int {
	return 1
}

func (q *QuestionBar) Size() (int, int) {
	return q.Width(), 1
}

func (q *QuestionBar) Render(updateScreen bool) {
	if !q.IsVisible() {
		return
	}

	q.input.Render(false)

	_, y := q.Position()

	components.RenderText(0, y, q.prompt, DefStyle)

	if updateScreen {
		screen.Show()
	}
}

func (q *QuestionBar) HandleEvent(ev tcell.Event) bool {
	if q.IsActive() {
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				q.closeBar()
				return true
			case tcell.KeyEnter:
				q.closeBar()
				q.enterFunc(q.input.Content())
			}
		}
	}

	return q.PanelImpl.HandleEvent(ev)
}

func (q *QuestionBar) closeBar() {
	q.Hide()
	components.Remove(q)
	window.Render()
}
