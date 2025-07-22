package ui

import (
	//"fmt"
	"fmt"
	"log"

	//"github.com/claude42/infiltrator/model"
	//"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

const textEntryPanelDefaultName = "This should not be seen here"

type TextEntryPanel struct {
	name  string
	y     int
	width int
	input Input

	ComponentImpl
}

func NewTextEntryPanel() *TextEntryPanel {
	t := &TextEntryPanel{name: textEntryPanelDefaultName}
	t.input = NewInputField()

	return t
}

func (p *TextEntryPanel) GetHeight() int {
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
	p.input.Render(updateScreen)

	if updateScreen {
		screen.Show()
	}
}

func (p *TextEntryPanel) determinePanelStyle() tcell.Style {
	if p.IsActive() {
		return ActivePanelStyle
	} else {
		return PanelStyle
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
	log.Println(".")
	p.input.SetContent(content)
	log.Println(".")
}

func (p *TextEntryPanel) HandleEvent(ev tcell.Event) bool {
	return p.input.HandleEvent(ev)
}

func (p *TextEntryPanel) SetActive(active bool) {
	p.ComponentImpl.SetActive(active)
	p.input.SetActive(active)
}
