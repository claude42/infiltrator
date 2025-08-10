package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
	"github.com/gdamore/tcell/v2"
)

type DateFilterPanel struct {
	PanelImpl

	start      *FilterInput
	end        *FilterInput
	lastActive *FilterInput
}

func NewDateFilterPanel() *DateFilterPanel {
	return &DateFilterPanel{
		PanelImpl: *NewPanelImpl(),
		start:     NewFilterInput(filter.DateFilterStart),
		end:       NewFilterInput(filter.DateFilterEnd),
	}
}

func (d *DateFilterPanel) Resize(x, y, width, height int) {
	d.PanelImpl.Resize(x, y, width, height)

	d.start.Resize(x+26, y, 20, 1)

	d.end.Resize(x+60, y, 20, 1)
}

func (d *DateFilterPanel) Render(updateScreen bool) {
	style := d.determinePanelStyle()

	header := fmt.Sprintf(" %s", d.name)
	x := renderText(0, d.y, header, style.Reverse(true))
	drawChars(x, d.y, d.width-(len(d.name)+1), ' ', style.Reverse(true))

	if d.start != nil {
		d.start.Render(updateScreen)
	}

	if d.end != nil {
		d.end.Render(updateScreen)
	}

	if updateScreen {
		screen.Show()
	}

}

func (d *DateFilterPanel) SetColorIndex(colorIndex uint8) {
	d.PanelImpl.SetColorIndex(colorIndex)

	d.start.SetColorIndex(colorIndex)
	d.end.SetColorIndex(colorIndex)

	if d.filter != nil {
		model.GetFilterManager().UpdateFilterColorIndex(d.filter, colorIndex)
	}
}

func (d *DateFilterPanel) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyTab:
			if d.start.IsActive() {
				d.start.SetActive(false)
				d.end.SetActive(true)
				return true
			}
		case tcell.KeyBacktab:
			if d.end.IsActive() {
				d.start.SetActive(true)
				d.end.SetActive(false)
				return true
			}
		}
	}

	if d.start.IsActive() && d.start.HandleEvent(ev) {
		return true
	}

	if d.end.IsActive() && d.end.HandleEvent(ev) {
		return true
	}

	return false
}

func (d *DateFilterPanel) SetActive(active bool) {
	d.PanelImpl.SetActive(active)

	if active && !d.start.IsActive() && !d.end.IsActive() {
		if d.lastActive != nil {
			d.lastActive.SetActive(true)
		} else {
			d.start.SetActive(true)
		}
	} else if !active {
		if d.start.IsActive() {
			d.lastActive = d.start
		} else if d.end.IsActive() {
			d.lastActive = d.end
		} else {
			d.lastActive = nil
		}
		d.start.SetActive(false)
		d.end.SetActive(false)
	}
}

func (d *DateFilterPanel) SetFilter(filter filter.Filter) {
	d.PanelImpl.SetFilter(filter)

	d.start.SetFilter(filter)
	d.end.SetFilter(filter)
}
