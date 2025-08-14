package components

import (
	"fmt"

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

	filter filter.Filter

	OldStyler     Styler
	CurrentStyler Styler
}

func NewPanelImpl(name string) *PanelImpl {
	p := &PanelImpl{
		name: name,
	}

	p.StyleUsing(p)

	return p
}

func (p *PanelImpl) Resize(x, y, width, height int) {
	// x, height get ignored
	p.ContainerImpl.Resize(0, y, width, 1)
}

func (p *PanelImpl) Height() int {
	return 1
}

func (p *PanelImpl) Size() (int, int) {
	return p.width, 1
}

func (t *PanelImpl) Render(updateScreen bool) {
	style := t.CurrentStyler.Style()

	header := fmt.Sprintf(" %s", t.name)
	RenderText(0, t.y, header, style.Reverse(true))

	if updateScreen {
		Screen.Show()
	}
}

func (p *PanelImpl) Style() tcell.Style {
	if p.IsActive() {
		return tcell.StyleDefault.Bold(true)
	} else {
		return tcell.StyleDefault
	}
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

func (p *PanelImpl) Name() string {
	return p.name
}

func (p *PanelImpl) StyleUsing(styler Styler) {
	if p.CurrentStyler != nil {
		p.OldStyler = p.CurrentStyler
	}
	p.CurrentStyler = styler
}
