package ui

import (
	"log"
	"slices"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
)

func setupNewStringFilterPanel(fn filter.StringFilterFuncFactory, name string,
	panelConfig *config.PanelTable) *StringFilterPanel {
	p := NewStringFilterPanel(name)
	f := filter.NewStringFilter(fn, p.Mode())
	model.GetFilterManager().AddFilter(f)
	p.SetFilter(f)

	// done last so both panel and filter get the same color index
	colorIndex := GetColorManager().Add(p)
	p.SetColorIndex(colorIndex)

	if panelConfig != nil {
		p.SetContent(panelConfig.Key)
		p.SetMode(config.FilterMode(slices.Index(config.FilterModeStrings, panelConfig.Mode)))
		p.SetCaseSensitive(panelConfig.CaseSensitive)
	}

	return p
}

func setupNewDateFilterPanel(name string, panelConfig *config.PanelTable) *DateFilterPanel {
	p := NewDateFilterPanel(name)
	filter := filter.NewDateFilter()
	model.GetFilterManager().AddFilter(filter)
	p.SetFilter(filter)

	colorIndex := GetColorManager().Add(p)
	p.SetColorIndex(colorIndex)

	if panelConfig != nil {
		p.SetTo(panelConfig.To)
		p.SetFrom(panelConfig.From)
	}

	return p
}

func NewPanel(panelType config.FilterType) components.Panel {
	return NewPanelWithPanelTypeAndConfig(panelType, nil)
}

func NewPanelWithConfig(panelConfig *config.PanelTable) components.Panel {
	panelType, err := config.FilterNameToType(panelConfig.Type)
	if err != nil {
		// TODO error handling
		return nil
	}

	return NewPanelWithPanelTypeAndConfig(panelType, panelConfig)
}

func NewPanelWithPanelTypeAndConfig(panelType config.FilterType,
	panelConfig *config.PanelTable) components.Panel {

	switch panelType {
	case config.FilterTypeKeyword:
		return setupNewStringFilterPanel(filter.DefaultStringFilterFuncFactory,
			config.Filters[panelType], panelConfig)
	case config.FilterTypeRegex:
		return setupNewStringFilterPanel(filter.RegexFilterFuncFactory,
			config.Filters[panelType], panelConfig)
	// case Glob:
	// 	return NewGlobPanel()
	// case Host:
	// 	return NewHostPanel()
	// case Facility:
	// 	return NewFacilityPanel()
	case config.FilterTypeDate:
		// TODO: error handling
		return setupNewDateFilterPanel(config.Filters[panelType], panelConfig)
	default:
		// TODO error handling: really panic here?
		log.Panicf("NewPanel() called with unknown panel type: %d",
			panelType)

		return nil
	}
}

func DestroyPanel(panel components.Panel) {
	fail.IfNil(panel, "DestroyPanel() called with nil panel")

	fm := model.GetFilterManager()
	fail.IfNil(fm, "DestroyPanel() called with nil pipeline")

	fm.RemoveFilter(panel.Filter())
	GetColorManager().Remove(panel)
}
