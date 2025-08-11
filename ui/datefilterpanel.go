package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
	"github.com/gdamore/tcell/v2"
)

type DateFilterPanel struct {
	PanelImpl

	from       *FilterInput
	to         *FilterInput
	lastActive *FilterInput
}

func NewDateFilterPanel(name string) *DateFilterPanel {
	return &DateFilterPanel{
		PanelImpl: *NewPanelImpl(name),
		from:      NewFilterInput(filter.DateFilterFrom),
		to:        NewFilterInput(filter.DateFilterTo),
	}
}

func (d *DateFilterPanel) Resize(x, y, width, height int) {
	d.PanelImpl.Resize(x, y, width, height)

	d.from.Resize(x+26, y, 20, 1)

	d.to.Resize(x+60, y, 20, 1)
}

func (d *DateFilterPanel) Render(updateScreen bool) {
	style := d.determinePanelStyle()

	header := fmt.Sprintf(" %s", d.name)
	x := renderText(0, d.y, header, style.Reverse(true))
	drawChars(x, d.y, d.width-(len(d.name)+1), ' ', style.Reverse(true))

	if d.from != nil {
		d.from.Render(updateScreen)
	}

	if d.to != nil {
		d.to.Render(updateScreen)
	}

	if updateScreen {
		screen.Show()
	}

}

func (d *DateFilterPanel) SetColorIndex(colorIndex uint8) {
	d.PanelImpl.SetColorIndex(colorIndex)

	d.from.SetColorIndex(colorIndex)
	d.to.SetColorIndex(colorIndex)

	if d.filter != nil {
		model.GetFilterManager().UpdateFilterColorIndex(d.filter, colorIndex)
	}
}

func (d *DateFilterPanel) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyTab:
			if d.from.IsActive() {
				d.from.SetActive(false)
				d.to.SetActive(true)
				return true
			}
		case tcell.KeyBacktab:
			if d.to.IsActive() {
				d.from.SetActive(true)
				d.to.SetActive(false)
				return true
			}
		}
	}

	if d.from.IsActive() && d.from.HandleEvent(ev) {
		return true
	}

	if d.to.IsActive() && d.to.HandleEvent(ev) {
		return true
	}

	return false
}

func (d *DateFilterPanel) SetActive(active bool) {
	d.PanelImpl.SetActive(active)

	if active && !d.from.IsActive() && !d.to.IsActive() {
		if d.lastActive != nil {
			d.lastActive.SetActive(true)
		} else {
			d.from.SetActive(true)
		}
	} else if !active {
		if d.from.IsActive() {
			d.lastActive = d.from
		} else if d.to.IsActive() {
			d.lastActive = d.to
		} else {
			d.lastActive = nil
		}
		d.from.SetActive(false)
		d.to.SetActive(false)
	}
}

func (d *DateFilterPanel) SetFilter(filter filter.Filter) {
	d.PanelImpl.SetFilter(filter)

	d.from.SetFilter(filter)
	d.to.SetFilter(filter)
}
