package ui

import (
	"slices"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"

	"github.com/gdamore/tcell/v2"
)

type StringFilterPanel struct {
	*FilterPanelImpl

	input         *FilterInput
	typeSelect    *ColoredDropdown
	mode          *ColoredDropdown
	caseSensitive *ColoredDropdown
}

func NewStringFilterPanel(panelType config.FilterType, name string) *StringFilterPanel {

	s := &StringFilterPanel{
		FilterPanelImpl: NewFilterPanelImpl(panelType, name),
		input:           NewFilterInput(name),
	}
	s.typeSelect = NewColoredDropdown(config.Filters.AllStrings(), tcell.KeyCtrlH, s.changePanelType)
	s.typeSelect.SetSelectedIndex(int(panelType))
	s.mode = NewColoredDropdown(config.FilterModeStrings, tcell.KeyCtrlJ, s.toggleMode)
	s.caseSensitive = NewColoredDropdown(config.CaseSensitiveStrings, tcell.KeyCtrlK, s.toggleCaseSensitive)
	s.Add(s.typeSelect)
	s.Add(s.mode)
	s.Add(s.caseSensitive)
	s.Add(s.input)

	return s
}

func (s *StringFilterPanel) SetPanelConfig(panelConfig *config.PanelTable) {
	if panelConfig == nil {
		return
	}

	s.SetContent(panelConfig.Key)
	mode := slices.Index(config.FilterModeStrings, panelConfig.Mode)
	if mode != -1 {
		s.SetMode(config.FilterMode(mode))
	}
	s.SetCaseSensitive(panelConfig.CaseSensitive)

	// don't put this into FilterPanelImpl!
	s.SetColorIndex(panelConfig.ColorIndex)
}

func (s *StringFilterPanel) Resize(x, y, width, height int) {
	s.FilterPanelImpl.Resize(x, y, width, height)

	s.typeSelect.Resize(x+1, y, config.PanelNameWidth, 1)
	s.input.Resize(x+config.PanelHeaderWidth+config.PanelHeaderGap, y,
		width-(x+config.PanelHeaderWidth+config.PanelHeaderGap), 1)
	s.mode.Resize(x+config.PanelNameWidth, y, 1, 1)
	s.caseSensitive.Resize(x+config.PanelNameWidth+8, y, 1, 1)
}

func (s *StringFilterPanel) Render(updateScreen bool) {
	if !s.IsVisible() {
		return
	}

	s.FilterPanelImpl.Render(false)

	style := s.CurrentStyler.Style()

	_, y := s.Position()
	components.RenderText(config.PanelHeaderWidth, y, "▶ ", style)

	if updateScreen {
		screen.Show()
	}
}

func (s *StringFilterPanel) SetColorIndex(colorIndex uint8) {
	s.FilterPanelImpl.SetColorIndex(colorIndex)

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
	s.FilterPanelImpl.SetFilter(filter)

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

	fail.IfNil(s.Filter(), "StringFilterPanel.SetMode() called without filter!")
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

	fail.IfNil(s.Filter(), "StringFilterPanel.SetCaseSensitive() called without filter!")
	model.GetFilterManager().UpdateFilterCaseSensitiveUpdate(s.Filter(), caseSensitive)
}

func (s *StringFilterPanel) SetName(name string) {
	s.FilterPanelImpl.SetName(name)
	s.input.SetName(name)
}

func (s *StringFilterPanel) changePanelType(i int) {
	newType := config.FilterType(i)
	if newType == s.panelType {
		return
	}

	s.panelConfig.Key = s.Content()
	s.panelConfig.Mode = s.Mode().String()
	s.panelConfig.CaseSensitive = s.CaseSensitive()
	s.panelConfig.ColorIndex = s.ColorIndex()

	// Note-to-self: don't put the next lines into FilterPanelImpl!
	newPanel := NewPanelWithPanelTypeAndConfig(newType, &s.panelConfig)
	newPanel.Show()
	err := window.ReplacePanel(s, newPanel)
	fail.OnError(err, "failed to replace panel")
}
