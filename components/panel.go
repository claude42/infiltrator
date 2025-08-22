package components

import (
	"github.com/claude42/infiltrator/model/filter"
	"github.com/gdamore/tcell/v2"
)

type Panel interface {
	Container
	Styler

	SetFilter(filter filter.Filter)
	Filter() filter.Filter
	SetName(name string)
	Name() string
}

type PanelImpl struct {
	ContainerImpl

	name string

	minHeight int

	filter filter.Filter

	OldStyler     Styler
	CurrentStyler Styler
}

func NewPanelImpl(name string) *PanelImpl {
	p := &PanelImpl{
		name:      name,
		minHeight: 1,
	}

	p.StyleUsing(p)

	return p
}

func (p *PanelImpl) SetMinHeight(minHeight int) {
	p.minHeight = minHeight
}

func (p *PanelImpl) Resize(x, y, width, height int) {
	// x, height get ignored
	p.ContainerImpl.Resize(0, y, width, max(p.minHeight, height))
}

func (p *PanelImpl) Height() int {
	return max(p.minHeight, p.height)
}

func (p *PanelImpl) Size() (int, int) {
	return p.width, p.Height()
}

func (p *PanelImpl) Style() tcell.Style {
	if p.IsActive() {
		return tcell.StyleDefault.Bold(true)
	} else {
		return tcell.StyleDefault
	}
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

func (p *PanelImpl) Name() string {
	return p.name
}

func (p *PanelImpl) StyleUsing(styler Styler) {
	if p.CurrentStyler != nil {
		p.OldStyler = p.CurrentStyler
	}
	p.CurrentStyler = styler
}
