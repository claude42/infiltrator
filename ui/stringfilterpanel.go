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
	panelType     *ColoredDropdown
	mode          *ColoredDropdown
	caseSensitive *ColoredDropdown
}

func NewStringFilterPanel(name string) *StringFilterPanel {
	s := &StringFilterPanel{
		ColoredPanel: NewColoredPanel(name),
		input:        NewFilterInput(name),
	}
	s.panelType = NewColoredDropdown(config.Filters.AllStrings(), tcell.KeyCtrlH, s.changePanelType)
	s.mode = NewColoredDropdown(config.FilterModeStrings, tcell.KeyCtrlJ, s.toggleMode)
	s.caseSensitive = NewColoredDropdown(config.CaseSensitiveStrings, tcell.KeyCtrlK, s.toggleCaseSensitive)
	s.Add(s.panelType)
	s.Add(s.mode)
	s.Add(s.caseSensitive)
	s.Add(s.input)

	return s
}

func (s *StringFilterPanel) Resize(x, y, width, height int) {
	s.ColoredPanel.Resize(x, y, width, height)

	s.input.Resize(x+headerWidth+2, y, width-(x+headerWidth+2), 1)
	s.mode.Resize(x+nameWidth, y, 1, 1)
	s.caseSensitive.Resize(x+nameWidth+8, y, 1, 1)
}

func (s *StringFilterPanel) Render(updateScreen bool) {
	if !s.IsVisible() {
		return
	}

	s.ColoredPanel.Render(false)

	style := s.CurrentStyler.Style()

	header := fmt.Sprintf(" %s", s.Name())
	_, y := s.Position()
	_ = components.RenderText(0, y, header, style.Reverse(true))
	components.RenderText(headerWidth, y, "▶ ", style)

	if updateScreen {
		screen.Show()
	}
}

func (s *StringFilterPanel) ColorIndex() uint8 {
	return s.colorIndex
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

func (s *StringFilterPanel) SetContent(content string) {
	fail.IfNil(s.input, "StringFilterPanel.SetContent() called without input field!")

	s.input.SetContent(content)
}

func (s *StringFilterPanel) Content() string {
	return s.input.Content()
}

func (s *StringFilterPanel) SetFilter(filter filter.Filter) {
	s.ColoredPanel.SetFilter(filter)

	s.input.SetFilter(filter)
}

func (s *StringFilterPanel) toggleMode(i int) {
	model.GetFilterManager().UpdateFilterMode(s.Filter(), config.FilterMode(i))

	s.Render(true)
}

func (s *StringFilterPanel) toggleCaseSensitive(i int) {
	model.GetFilterManager().UpdateFilterCaseSensitiveUpdate(s.Filter(), i != 0)

	s.Render(true)
}

func (s *StringFilterPanel) Mode() config.FilterMode {
	return config.FilterMode(s.mode.SelectedIndex())
}

func (s *StringFilterPanel) SetMode(mode config.FilterMode) {
	s.mode.SetSelectedIndex(int(mode))
	model.GetFilterManager().UpdateFilterMode(s.Filter(), mode)
}

func (s *StringFilterPanel) CaseSensitive() bool {
	return s.caseSensitive.SelectedIndex() == 1
}

func (s *StringFilterPanel) SetCaseSensitive(caseSensitive bool) {

	if caseSensitive {
		s.caseSensitive.SetSelectedIndex(1)
	} else {
		s.caseSensitive.SetSelectedIndex(0)
	}

	model.GetFilterManager().UpdateFilterCaseSensitiveUpdate(s.Filter(), caseSensitive)

}

func (s *StringFilterPanel) SetName(name string) {
	s.ColoredPanel.SetName(name)
	s.input.SetName(name)
}

func (s *StringFilterPanel) changePanelType(i int) {
	filterType := config.Filters[i].FilterType
	filterString := config.Filters[i].FilterString

	newPanel := NewPanel(filterType)
	switch newPanel := newPanel.(type) {
	case *StringFilterPanel:
		newPanel.SetContent(s.Content())
		newPanel.SetMode(s.Mode())
		newPanel.SetCaseSensitive(s.CaseSensitive())
		newPanel.SetColorIndex(s.ColorIndex())
	case *DateFilterPanel:
		// can't transfer content
		// can't transfer mode
		// can't transfer case sensitivity
	}

	newPanel.SetColorIndex(s.ColorIndex())
	newPanel.SetMode(s.Mode())
	newPanel.SetCaseSensitive(s.CaseSensitive())

	var fn filter.StringFilterFuncFactory
	switch filterType {
	case config.FilterTypeKeyword:
		fn = filter.DefaultStringFilterFuncFactory
	case config.FilterTypeRegex:
		fn = filter.NewRegexFilterFuncFactory()
	default:
		// TODO error handling
		return
	}

	model.GetFilterManager().ChangeStringFilterType(s.Filter(), fn, filterType, s.Mode())

	s.Render(true)
}
