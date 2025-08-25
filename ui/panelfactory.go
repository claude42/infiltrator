package ui

import (
	"log"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
)

func setupNewStringFilterPanel(panelType config.FilterType,
	fn filter.StringFilterFuncFactory, name string,
	panelConfig *config.PanelTable) *StringFilterPanel {

	p := NewStringFilterPanel(panelType, name)
	f := filter.NewStringFilter(fn, p.Mode())
	model.GetFilterManager().AddFilter(f)
	p.SetFilter(f)

	if panelConfig != nil {
		p.SetPanelConfig(panelConfig)
	}
	// done last so both panel and filter get the same color index
	if panelConfig == nil || panelConfig.ColorIndex == 0 {
		colorIndex := GetColorManager().Add(p)
		p.SetColorIndex(colorIndex)
	}

	return p
}

func setupNewDateFilterPanel(panelType config.FilterType,
	name string, panelConfig *config.PanelTable) *DateFilterPanel {

	p := NewDateFilterPanel(panelType, name)
	filter := filter.NewDateFilter()
	model.GetFilterManager().AddFilter(filter)
	p.SetFilter(filter)

	if panelConfig != nil {
		p.SetPanelConfig(panelConfig)
	}
	// done last so both panel and filter get the same color index
	if panelConfig == nil || panelConfig.ColorIndex == 0 {
		colorIndex := GetColorManager().Add(p)
		p.SetColorIndex(colorIndex)
	}

	return p
}

func NewPanel(panelType config.FilterType) FilterPanel {
	return NewPanelWithPanelTypeAndConfig(panelType, nil)
}

func NewPanelWithConfig(panelConfig *config.PanelTable) FilterPanel {
	panelType, err := config.Filters.Type(panelConfig.Type)
	if err != nil {
		// TODO error handling
		return nil
	}

	return NewPanelWithPanelTypeAndConfig(panelType, panelConfig)
}

func NewPanelWithPanelTypeAndConfig(panelType config.FilterType,
	panelConfig *config.PanelTable) FilterPanel {

	filterString, err := config.Filters.String(panelType)
	fail.OnError(err, "error getting filter string for filter type")
	switch panelType {
	case config.FilterTypeKeyword:
		return setupNewStringFilterPanel(panelType, filter.DefaultStringFilterFuncFactory,
			filterString, panelConfig)
	case config.FilterTypeRegex:
		return setupNewStringFilterPanel(panelType, filter.RegexFilterFuncFactory,
			filterString, panelConfig)
	// case Glob:
	// 	return NewGlobPanel()
	// case Host:
	// 	return NewHostPanel()
	// case Facility:
	// 	return NewFacilityPanel()
	case config.FilterTypeDate:
		// TODO: error handling
		return setupNewDateFilterPanel(panelType, filterString, panelConfig)
	default:
		// TODO error handling: really panic here?
		log.Panicf("NewPanel() called with unknown panel type: %d",
			panelType)

		return nil
	}
}

func DestroyPanel(panel FilterPanel) {
	fail.IfNil(panel, "DestroyPanel() called with nil panel")

	fm := model.GetFilterManager()
	fail.IfNil(fm, "DestroyPanel() called with nil pipeline")

	fm.RemoveFilter(panel.Filter())
	GetColorManager().Remove(panel)
}
