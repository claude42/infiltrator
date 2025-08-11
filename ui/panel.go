package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/model/filter"
	"github.com/gdamore/tcell/v2"
	//"github.com/claude42/infiltrator/util"
	// "github.com/gdamore/tcell/v2"
)

type Panel interface {
	Component

	Height() int
	Position() (int, int)
	SetColorIndex(colorIndex uint8)
	SetFilter(filter filter.Filter)
	Filter() filter.Filter
	SetName(name string)
}

type PanelImpl struct {
	ComponentImpl

	name       string
	y          int
	width      int
	colorIndex uint8
	filter     filter.Filter
}

func NewPanelImpl(name string) *PanelImpl {
	return &PanelImpl{
		name: name,
	}
}

func (p *PanelImpl) Position() (int, int) {
	return 0, p.y
}

func (p *PanelImpl) Height() int {
	return 1
}

func (p *PanelImpl) Resize(x, y, width, height int) {
	// x, height get ignored
	p.y = y
	p.width = width
}

func (t *PanelImpl) Render(updateScreen bool) {
	style := t.determinePanelStyle()

	header := fmt.Sprintf(" %s", t.name)
	renderText(0, t.y, header, style.Reverse(true))

	if updateScreen {
		screen.Show()
	}
}

func (p *PanelImpl) determinePanelStyle() tcell.Style {
	if p.IsActive() {
		return tcell.StyleDefault.Bold(true).Foreground(FilterColors[p.colorIndex][0])
	} else {
		return tcell.StyleDefault.Foreground(FilterColors[p.colorIndex][1])
	}
}

func (p *PanelImpl) SetColorIndex(colorIndex uint8) {
	p.colorIndex = colorIndex
}

// ignore all events
func (p *PanelImpl) HandleEvent(ev tcell.Event) bool {
	return false
}

func (p *PanelImpl) SetFilter(filter filter.Filter) {
	p.filter = filter
}

func (p *PanelImpl) Filter() filter.Filter {
	return p.filter
}

func (p *PanelImpl) SetName(name string) {
	p.name = name
}
