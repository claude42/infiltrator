package ui

import (
	//"fmt"
	"fmt"
	"log"

	"github.com/claude42/infiltrator/model"
	//"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

const textEntryPanelDefaultName = "This should not be seen here"

type TextEntryPanel struct {
	name       string
	y          int
	width      int
	input      Input
	colorIndex uint8
	filter     model.Filter

	ComponentImpl
}

func NewTextEntryPanel() *TextEntryPanel {
	t := &TextEntryPanel{name: textEntryPanelDefaultName}
	t.input = NewInputField()

	return t
}

func (p *TextEntryPanel) Height() int {
	return 2
}

func (p *TextEntryPanel) Resize(x, y, width, height int) {
	// x, height get ignored
	p.y = y
	p.width = width
	p.input.Resize(x+2, y+1, width-4, 1)
}

func (p *TextEntryPanel) Render(updateScreen bool) {
	style := p.determinePanelStyle()
	p.renderHeadline(style)
	p.renderContent(style)
	if p.input != nil {
		p.input.Render(updateScreen)
	}

	if updateScreen {
		screen.Show()
	}
}

func (p *TextEntryPanel) SetColorIndex(colorIndex uint8) {
	p.colorIndex = colorIndex
	if p.filter != nil {
		p.filter.SetColorIndex(colorIndex)
	}
}

func (p *TextEntryPanel) determinePanelStyle() tcell.Style {
	if p.IsActive() {
		return tcell.StyleDefault.Bold(true).Background(FilterColors[p.colorIndex][0])
	} else {
		return tcell.StyleDefault.Background(FilterColors[p.colorIndex][0])
	}
}

func (p *TextEntryPanel) renderHeadline(style tcell.Style) {
	headline := fmt.Sprintf("─ %s ", p.name)
	x := renderText(0, p.y, headline, style)
	fillChars(x, p.y, '─', style)
}

func (p *TextEntryPanel) renderContent(style tcell.Style) {
	fillChars(0, p.y+1, ' ', style)
}

func (p *TextEntryPanel) SetContent(content string) {
	if p.input == nil {
		log.Panicln("TextEntryPanel.SetContent() called without input field!")
		return
	}
	p.input.SetContent(content)

}

func (p *TextEntryPanel) HandleEvent(ev tcell.Event) bool {
	if p.input == nil {
		return false
	}
	return p.input.HandleEvent(ev)
}

func (p *TextEntryPanel) SetActive(active bool) {
	p.ComponentImpl.SetActive(active)
	p.input.SetActive(active)
}

func (p *TextEntryPanel) SetName(name string) {
	p.name = name
}

func (p *TextEntryPanel) SetFilter(filter model.Filter) {
	p.filter = filter

	if p.input == nil {
		log.Panicln("TextEntryPanel.SetFilter() called without input field!")
		return
	}
	p.input.Watch(filter)
}

func (p *TextEntryPanel) Filter() model.Filter {
	return p.filter
}
