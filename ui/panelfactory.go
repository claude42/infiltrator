package ui

import (
	"log"

	"github.com/claude42/infiltrator/model"
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

func setupNewTinyPanel(fn model.StringFilterFuncFactory, name string) (*TinyPanel, error) {
	p := NewTinyPanel()
	p.SetName(name)
	filter := model.NewStringFilter(fn, p.Mode())
	model.GetFilterManager().AddFilter(filter)
	p.SetFilter(filter)

	// done last so both panel and filter get the same color index
	colorIndex, err := GetColorManager().Add(p)
	if err != nil {
		return nil, err
	}
	p.SetColorIndex(colorIndex)

	return p, nil
}

func NewPanel(panelType PanelType) (Panel, error) {
	switch panelType {
	case PanelTypeKeyword:
		return setupNewTinyPanel(model.DefaultStringFilterFuncFactory, keywordPanelDefaultName)
		// return createNewKeywordPanel()
	case PanelTypeRegex:
		return setupNewTinyPanel(model.RegexFilterFuncFactory, regexPanelDefaultName)
		// return createNewRegexPanel()
	/*case Glob:
		return NewGlobPanel()
	case Host:
		return NewHostPanel()
	case Facility:
		return NewFacilityPanel()
	case Date:
		return NewDatePanel()*/
	default:
		log.Panicln("NewPanel() called with unknown panel type:", panelType)
		return nil, nil
	}
}

func DestroyPanel(panel Panel) {
	if panel == nil {
		log.Panicln("DestroyPanel() called with nil panel")
		return
	}

	fm := model.GetFilterManager()
	if fm == nil {
		log.Panicln("DestroyPanel() called with nil pipeline")
		return
	}

	fm.RemoveFilter(panel.Filter())
	GetColorManager().Remove(panel)
}
