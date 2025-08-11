package ui

import (
	"log"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
	// "github.com/gdamore/tcell/v2"
)

func setupNewStringFilterPanel(fn filter.StringFilterFuncFactory, name string) *StringFilterPanel {
	p := NewStringFilterPanel(name)
	filter := filter.NewStringFilter(fn, p.Mode())
	model.GetFilterManager().AddFilter(filter)
	p.SetFilter(filter)

	// done last so both panel and filter get the same color index
	colorIndex := GetColorManager().Add(p)

	p.SetColorIndex(colorIndex)

	return p
}

func setupNewDateFilterPanel() *DateFilterPanel {
	p := NewDateFilterPanel(config.Filters[config.FilterTypeDate])
	filter := filter.NewDateFilter()
	model.GetFilterManager().AddFilter(filter)
	p.SetFilter(filter)

	return p
}

func NewPanel(panelType config.FilterType) Panel {
	switch panelType {
	case config.FilterTypeKeyword:
		return setupNewStringFilterPanel(filter.DefaultStringFilterFuncFactory,
			config.Filters[config.FilterTypeKeyword])
	case config.FilterTypeRegex:
		return setupNewStringFilterPanel(filter.RegexFilterFuncFactory,
			config.Filters[config.FilterTypeRegex])
	// case Glob:
	// 	return NewGlobPanel()
	// case Host:
	// 	return NewHostPanel()
	// case Facility:
	// 	return NewFacilityPanel()
	case config.FilterTypeDate:
		// TODO: error handling
		return setupNewDateFilterPanel()
	default:
		log.Panicln("NewPanel() called with unknown panel type:", panelType)
		return nil
	}
}

func DestroyPanel(panel Panel) {
	fail.IfNil(panel, "DestroyPanel() called with nil panel")

	fm := model.GetFilterManager()
	fail.IfNil(fm, "DestroyPanel() called with nil pipeline")

	fm.RemoveFilter(panel.Filter())
	GetColorManager().Remove(panel)
}
