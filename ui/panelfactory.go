package ui

import (
	"log"

	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
	// "github.com/gdamore/tcell/v2"
)

type PanelType int

const (
	PanelTypeKeyword PanelType = iota
	PanelTypeRegex
	PanelTypeGlob
	PanelTypeHost
	PanelTypeFacility
	PanelTypeDate
)

const keywordPanelDefaultName = "Keyword"
const regexPanelDefaultName = "Regex"
const dateFilterDefaultName = "Date"

func setupNewStringFilterPanel(fn filter.StringFilterFuncFactory, name string) *StringFilterPanel {
	p := NewStringFilterPanel()
	p.SetName(name)
	filter := filter.NewStringFilter(fn, p.Mode())
	model.GetFilterManager().AddFilter(filter)
	p.SetFilter(filter)

	// done last so both panel and filter get the same color index
	colorIndex := GetColorManager().Add(p)

	p.SetColorIndex(colorIndex)

	return p
}

func setupNewDateFilterPanel() *DateFilterPanel {
	p := NewDateFilterPanel()
	p.SetName(dateFilterDefaultName)
	filter := filter.NewDateFilter()
	model.GetFilterManager().AddFilter(filter)
	p.SetFilter(filter)

	return p
}

func NewPanel(panelType PanelType) Panel {
	switch panelType {
	case PanelTypeKeyword:
		return setupNewStringFilterPanel(filter.DefaultStringFilterFuncFactory, keywordPanelDefaultName)
	case PanelTypeRegex:
		return setupNewStringFilterPanel(filter.RegexFilterFuncFactory, regexPanelDefaultName)
	// case Glob:
	// 	return NewGlobPanel()
	// case Host:
	// 	return NewHostPanel()
	// case Facility:
	// 	return NewFacilityPanel()
	case PanelTypeDate:
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
