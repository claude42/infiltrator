package ui

import (
	"log"

	"github.com/claude42/infiltrator/model"
	// "github.com/gdamore/tcell/v2"
)

const (
	TypeKeyword int = iota
	TypeRegex
	Glob
	Host
	Facility
	Date
)

var FilterModes = []string{
	"match",
	"focus",
	"hide",
}

const keywordPanelDefaultName = "Keyword"
const regexPanelDefaultName = "Regex"

func setupNewTinyPanel(fn model.StringFilterFuncFactory, name string, mode int) (*TinyPanel, error) {
	p := NewTinyPanel(mode)
	p.SetName(name)
	filter := model.NewStringFilter(fn, mode)
	model.GetPipeline().AddFilter(filter)
	p.SetFilter(filter)
	p.WatchInput(filter)

	// done last so both panel and filter get the same color index
	colorIndex, err := GetColorManager().Add(p)
	if err != nil {
		return nil, err
	}
	p.SetColorIndex(colorIndex)

	return p, nil
}

func NewPanel(panelType int, mode int) (Panel, error) {
	switch panelType {
	case TypeKeyword:
		return setupNewTinyPanel(model.DefaultStringFilterFuncFactory, keywordPanelDefaultName, mode)
		// return createNewKeywordPanel()
	case TypeRegex:
		return setupNewTinyPanel(model.RegexFilterFuncFactory, regexPanelDefaultName, mode)
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

	pipeline := model.GetPipeline()
	if pipeline == nil {
		log.Panicln("DestroyPanel() called with nil pipeline")
		return
	}

	pipeline.RemoveFilter(panel.Filter())
	GetColorManager().Remove(panel)
}
