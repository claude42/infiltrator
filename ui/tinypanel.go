package ui

import (
	//"fmt"
	"fmt"
	"log"

	"github.com/claude42/infiltrator/model"
	//"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

const tinyPanelDefaultName = "This should not be seen here"

const nameWidth = 9
const headerWidth = 24

type TinyPanel struct {
	name            string
	y               int
	width           int
	input           Input
	mode            *Select
	caseSensitivity *Select
	colorIndex      uint8
	filter          model.Filter

	ComponentImpl
}

func NewTinyPanel(mode int) *TinyPanel {
	t := &TinyPanel{name: tinyPanelDefaultName}
	t.input = NewInputField()
	t.mode = NewSelect(FilterModes)
	t.caseSensitivity = NewSelect([]string{"case", "cAsE"})

	return t
}

func (p *TinyPanel) Height() int {
	return 1
}

func (p *TinyPanel) Resize(x, y, width, height int) {
	// x, height get ignored
	p.y = y
	p.width = width
	p.input.Resize(x+headerWidth+2, y, width-len(p.name)-5, 1)
	p.mode.Resize(x+nameWidth, y, 1, 1)
	p.caseSensitivity.Resize(x+nameWidth+8, y, 1, 1)
}

func (p *TinyPanel) Render(updateScreen bool) {
	style := p.determinePanelStyle()

	header := fmt.Sprintf(" %s", p.name)
	x := renderText(0, p.y, header, style.Reverse(true))
	drawChars(x, p.y, headerWidth-x, ' ', style.Reverse((true)))
	renderText(headerWidth, p.y, "► ", style)

	if p.input != nil {
		p.input.Render(updateScreen)
	}

	if p.mode != nil {
		p.mode.Render(updateScreen)
	}

	if p.caseSensitivity != nil {
		p.caseSensitivity.Render(updateScreen)
	}

	if updateScreen {
		screen.Show()
	}
}

func (p *TinyPanel) SetColorIndex(colorIndex uint8) {
	p.colorIndex = colorIndex
	p.input.SetColorIndex(colorIndex)
	p.mode.SetColorIndex(colorIndex)
	p.caseSensitivity.SetColorIndex(colorIndex)
	if p.filter != nil {
		p.filter.SetColorIndex(colorIndex)
	}
}

func (p *TinyPanel) determinePanelStyle() tcell.Style {
	if p.IsActive() {
		return tcell.StyleDefault.Bold(true).Foreground(FilterColors[p.colorIndex][0])
	} else {
		return tcell.StyleDefault.Foreground(FilterColors[p.colorIndex][1])
	}
}

func (p *TinyPanel) SetContent(content string) {
	if p.input == nil {
		log.Panicln("TinyPanel.SetContent() called without input field!")
		return
	}
	p.input.SetContent(content)

}

func (p *TinyPanel) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlM:
			p.toggleMode()
		case tcell.KeyCtrlH:
			p.toggleCaseSensitivity()
		}
	}

	if p.input == nil {
		return false
	}
	return p.input.HandleEvent(ev)
}

func (p *TinyPanel) SetActive(active bool) {
	p.ComponentImpl.SetActive(active)
	p.input.SetActive(active)
	p.mode.SetActive(active)
	p.caseSensitivity.SetActive(active)
}

func (p *TinyPanel) SetName(name string) {
	p.name = name
}

func (p *TinyPanel) SetFilter(filter model.Filter) {
	p.filter = filter
}

func (p *TinyPanel) Filter() model.Filter {
	return p.filter
}

func (p *TinyPanel) WatchInput(eh tcell.EventHandler) {
	if p.input == nil {
		log.Panicln("TinyPanel.WatchInput() called without input field!")
		return
	}
	p.input.Watch(eh)
}

func (p *TinyPanel) toggleMode() {
	p.filter.SetMode(p.mode.NextOption())

	p.Render(true)
}

func (p *TinyPanel) toggleCaseSensitivity() {
	p.caseSensitivity.NextOption()

	p.Render(true)
}
