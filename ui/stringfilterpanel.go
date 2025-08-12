package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"

	"github.com/gdamore/tcell/v2"
)

const nameWidth = 9
const headerWidth = 24

var filterModes = []string{
	"focus",
	"match",
	"hide",
}

var caseSensitive = []string{
	"case",
	"CaSe",
}

type StringFilterPanel struct {
	*ColoredPanel

	input         *FilterInput
	mode          *ColoredSelect
	caseSensitive *ColoredSelect
}

func NewStringFilterPanel(name string) *StringFilterPanel {
	s := &StringFilterPanel{
		ColoredPanel:  NewColoredPanel(name),
		input:         NewFilterInput(name),
		mode:          NewColoredSelect(filterModes),
		caseSensitive: NewColoredSelect(caseSensitive),
	}

	return s
}

func (t *StringFilterPanel) Resize(x, y, width, height int) {
	t.ColoredPanel.Resize(x, y, width, height)

	t.input.Resize(x+headerWidth+2, y, width-(x+headerWidth+2), 1)
	t.mode.Resize(x+nameWidth, y, 1, 1)
	t.caseSensitive.Resize(x+nameWidth+8, y, 1, 1)
}

func (t *StringFilterPanel) Render(updateScreen bool) {
	style := t.ColoredPanel.CurrentStyler.Style()

	header := fmt.Sprintf(" %s", t.Name())
	_, y := t.Position()
	x := components.RenderText(0, y, header, style.Reverse(true))
	components.DrawChars(x, y, headerWidth-x, ' ', style.Reverse((true)))
	components.RenderText(headerWidth, y, "► ", style)

	if t.input != nil {
		t.input.Render(updateScreen)
	}

	if t.mode != nil {
		t.mode.Render(updateScreen)
	}

	if t.caseSensitive != nil {
		t.caseSensitive.Render(updateScreen)
	}

	if updateScreen {
		screen.Show()
	}
}

func (s *StringFilterPanel) SetColorIndex(colorIndex uint8) {
	s.ColoredPanel.SetColorIndex(colorIndex)

	s.input.SetColorIndex(colorIndex)
	s.mode.SetColorIndex(colorIndex)
	s.caseSensitive.SetColorIndex(colorIndex)
	if s.Filter() != nil {
		model.GetFilterManager().UpdateFilterColorIndex(s.Filter(), colorIndex)
	}
}

func (t *StringFilterPanel) SetContent(content string) {
	fail.IfNil(t.input, "StringFilterPanel.SetContent() called without input field!")

	t.input.SetContent(content)
}

func (t *StringFilterPanel) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlS:
			t.toggleMode()
			return true
		case tcell.KeyCtrlH:
			t.toggleCaseSensitive()
			return true
		}
	case *tcell.EventMouse:
		buttons := ev.Buttons()
		if buttons&tcell.ButtonPrimary != 0 {
			if t.mouseToggleMode(ev) {
				return true
			}
			if t.mouseToggleCaseSensitive(ev) {
				return true
			}

		}
	}

	if t.input != nil && t.input.HandleEvent(ev) {
		return true
	}

	return false
}

func (t *StringFilterPanel) SetActive(active bool) {
	t.ColoredPanel.SetActive(active)

	t.input.SetActive(active)
	t.mode.SetActive(active)
	t.caseSensitive.SetActive(active)
}

func (t *StringFilterPanel) SetFilter(filter filter.Filter) {
	t.ColoredPanel.SetFilter(filter)

	t.input.SetFilter(filter)
}

// func (t *StringFilterPanel) WatchInput(eh tcell.EventHandler) {
// 	if t.input == nil {
// 		log.Panicln("StringFilterPanel.WatchInput() called without input field!")
// 		return
// 	}
// 	t.input.Watch(eh)
// }

func (t *StringFilterPanel) toggleMode() {
	model.GetFilterManager().UpdateFilterMode(t.Filter(), filter.FilterMode(t.mode.NextOption()))

	t.Render(true)
}

func (t *StringFilterPanel) mouseToggleMode(ev *tcell.EventMouse) bool {
	mouseX, mouseY := ev.Position()
	modeX, modeY := t.mode.Position()
	if mouseX >= modeX && mouseX <= modeX+t.mode.Width() &&
		mouseY == modeY {
		t.toggleMode()
		return true
	} else {
		return false
	}
}

func (t *StringFilterPanel) toggleCaseSensitive() {
	model.GetFilterManager().UpdateFilterCaseSensitiveUpdate(t.Filter(), t.caseSensitive.NextOption() != 0)

	t.Render(true)
}

func (t *StringFilterPanel) mouseToggleCaseSensitive(ev *tcell.EventMouse) bool {
	mouseX, mouseY := ev.Position()
	caseSensitiveX, caseSensitiveY := t.caseSensitive.Position()
	if mouseX >= caseSensitiveX &&
		mouseX <= caseSensitiveX+t.caseSensitive.Width() &&
		mouseY == caseSensitiveY {
		t.toggleCaseSensitive()
		return true
	} else {
		return false
	}
}

func (t *StringFilterPanel) Mode() filter.FilterMode {
	return filter.FilterMode(t.mode.SelectedIndex())
}

func (t *StringFilterPanel) SetMode(mode int) {
	t.mode.SetSelectedIndex(mode)
}

func (t *StringFilterPanel) SetName(name string) {
	t.ColoredPanel.SetName(name)
	t.input.SetName(name)
}
