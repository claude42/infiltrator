package components

import (
	"fmt"
	"log"

	"github.com/claude42/infiltrator/model/filter"
	"github.com/gdamore/tcell/v2"
	//"github.com/claude42/infiltrator/util"
	// "github.com/gdamore/tcell/v2"
)

type Panel interface {
	Component
	Styler

	Height() int
	Width() int
	Position() (int, int)
	SetFilter(filter filter.Filter)
	Filter() filter.Filter
	SetName(name string)
	Name() string
}

type PanelImpl struct {
	ComponentImpl

	name  string
	y     int
	width int

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

func (p *PanelImpl) Position() (int, int) {
	return 0, p.y
}

func (p *PanelImpl) Height() int {
	return 1
}

func (p *PanelImpl) Width() int {
	return p.width
}

func (p *PanelImpl) Resize(x, y, width, height int) {
	// x, height get ignored
	p.y = y
	p.width = width
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
	log.Printf("Styler: %T\n%+v\n%p", styler, styler, styler)
	if p.CurrentStyler != nil {
		p.OldStyler = p.CurrentStyler
	}
	p.CurrentStyler = styler
}
