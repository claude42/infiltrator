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

const keywordPanelDefaultName = "Keyword"
const regexPanelDefaultName = "Regex"

func setupNewTextEntryPanel(fn model.StringFilterFuncFactory, name string) *TextEntryPanel {
	p := NewTextEntryPanel()
	p.SetName(name)
	filter := model.NewStringFilter(fn)
	model.GetPipeline().AddFilter(filter)
	p.SetFilter(filter)
	p.SetReceiver(filter)

	// done last so both panel and filter get the same color index
	p.SetColorIndex(GetColorManager().Add(p))

	return p
}

func NewPanel(panelType int) Panel {
	switch panelType {
	case TypeKeyword:
		return setupNewTextEntryPanel(model.DefaultStringFilterFuncFactory, keywordPanelDefaultName)
		// return createNewKeywordPanel()
	case TypeRegex:
		return setupNewTextEntryPanel(model.RegexFilterFuncFactory, regexPanelDefaultName)
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
		return nil
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
