package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"

	"github.com/gdamore/tcell/v2"
)

const nameWidth = 9
const headerWidth = 24

type StringFilterPanel struct {
	*ColoredPanel

	input         *FilterInput
	mode          *ColoredDropdown
	caseSensitive *ColoredDropdown
}

func NewStringFilterPanel(name string) *StringFilterPanel {
	s := &StringFilterPanel{
		ColoredPanel: NewColoredPanel(name),
		input:        NewFilterInput(name),
	}
	s.mode = NewColoredDropdown(config.FilterModeStrings, tcell.KeyCtrlS, s.toggleMode)
	s.caseSensitive = NewColoredDropdown(config.CaseSensitiveStrings, tcell.KeyCtrlH, s.toggleCaseSensitive)
	s.ColoredPanel.Add(s.mode)
	s.ColoredPanel.Add(s.caseSensitive)
	s.ColoredPanel.Add(s.input)

	return s
}

func (t *StringFilterPanel) Resize(x, y, width, height int) {
	t.ColoredPanel.Resize(x, y, width, height)

	t.input.Resize(x+headerWidth+2, y, width-(x+headerWidth+2), 1)
	t.mode.Resize(x+nameWidth, y, 1, 1)
	t.caseSensitive.Resize(x+nameWidth+8, y, 1, 1)
}

func (t *StringFilterPanel) Render(updateScreen bool) {
	if !t.IsVisible() {
		return
	}

	t.ColoredPanel.Render(false)

	style := t.ColoredPanel.CurrentStyler.Style()

	header := fmt.Sprintf(" %s", t.Name())
	_, y := t.Position()
	_ = components.RenderText(0, y, header, style.Reverse(true))
	components.RenderText(headerWidth, y, "▶ ", style)

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

func (t *StringFilterPanel) Content() string {
	return t.input.Content()
}

func (t *StringFilterPanel) SetFilter(filter filter.Filter) {
	t.ColoredPanel.SetFilter(filter)

	t.input.SetFilter(filter)
}

func (t *StringFilterPanel) toggleMode(i int) {
	model.GetFilterManager().UpdateFilterMode(t.Filter(), config.FilterMode(i))

	t.Render(true)
}

func (t *StringFilterPanel) toggleCaseSensitive(i int) {
	model.GetFilterManager().UpdateFilterCaseSensitiveUpdate(t.Filter(), i != 0)

	t.Render(true)
}

func (t *StringFilterPanel) Mode() config.FilterMode {
	return config.FilterMode(t.mode.SelectedIndex())
}

func (t *StringFilterPanel) SetMode(mode config.FilterMode) {
	t.mode.SetSelectedIndex(int(mode))
	model.GetFilterManager().UpdateFilterMode(t.Filter(), mode)
}

func (t *StringFilterPanel) CaseSensitive() bool {
	return t.caseSensitive.SelectedIndex() == 1
}

func (t *StringFilterPanel) SetCaseSensitive(caseSensitive bool) {

	if caseSensitive {
		t.caseSensitive.SetSelectedIndex(1)
	} else {
		t.caseSensitive.SetSelectedIndex(0)
	}

	model.GetFilterManager().UpdateFilterCaseSensitiveUpdate(t.Filter(), caseSensitive)

}

func (t *StringFilterPanel) SetName(name string) {
	t.ColoredPanel.SetName(name)
	t.input.SetName(name)
}
