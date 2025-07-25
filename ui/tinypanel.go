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
const headerWidth = 9

var modes = map[int]string{
	model.FilterMatch:     "[match] ",
	model.FilterHighlight: "[mark ] ",
	model.FilterHide:      "[hide ] ",
}

type TinyPanel struct {
	name       string
	y          int
	width      int
	input      Input
	colorIndex uint8
	filter     model.Filter
	mode       int

	ComponentImpl
}

func NewTinyPanel() *TinyPanel {
	t := &TinyPanel{name: tinyPanelDefaultName}
	t.input = NewInputField()

	return t
}

func (p *TinyPanel) Height() int {
	return 1
}

func (p *TinyPanel) Resize(x, y, width, height int) {
	// x, height get ignored
	p.y = y
	p.width = width
	p.input.Resize(x+headerWidth+7+2, y, width-len(p.name)-5, 1)
}

func (p *TinyPanel) Render(updateScreen bool) {
	style := p.determinePanelStyle()

	header := fmt.Sprintf(" %s", p.name)
	x := renderText(0, p.y, header, style.Reverse(true))
	x = drawChars(x, p.y, headerWidth-x, ' ', style.Reverse((true)))

	x = renderText(x, p.y, modes[p.mode], style.Reverse(true))

	screen.SetContent(x, p.y, '►', nil, style)
	screen.SetContent(x+1, p.y, ' ', nil, style)

	if p.input != nil {
		p.input.Render(updateScreen)
	}

	if updateScreen {
		screen.Show()
	}
}

func (p *TinyPanel) SetColorIndex(colorIndex uint8) {
	p.colorIndex = colorIndex
	p.input.SetColorIndex(colorIndex)
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
	if p.input == nil {
		return false
	}
	return p.input.HandleEvent(ev)
}

func (p *TinyPanel) SetActive(active bool) {
	p.ComponentImpl.SetActive(active)
	p.input.SetActive(active)
}

func (p *TinyPanel) SetName(name string) {
	p.name = name
}

func (p *TinyPanel) SetFilter(filter model.Filter) {
	p.filter = filter

	if p.input == nil {
		log.Panicln("TinyPanel.SetFilter() called without input field!")
		return
	}
	p.input.Watch(filter)
}

func (p *TinyPanel) Filter() model.Filter {
	return p.filter
}
