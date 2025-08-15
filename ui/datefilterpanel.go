package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
	"github.com/gdamore/tcell/v2"
)

type DateFilterPanel struct {
	*ColoredPanel

	from       *FilterInput
	to         *FilterInput
	lastActive *FilterInput
}

func NewDateFilterPanel(name string) *DateFilterPanel {
	d := &DateFilterPanel{
		ColoredPanel: NewColoredPanel(name),
		from:         NewFilterInput(filter.DateFilterFrom),
		to:           NewFilterInput(filter.DateFilterTo),
	}
	d.ColoredPanel.Add(d.from)
	d.ColoredPanel.Add(d.to)
	d.to.SetActive(false)
	return d
}

func (d *DateFilterPanel) Resize(x, y, width, height int) {
	d.ColoredPanel.Resize(x, y, width, height)

	d.from.Resize(x+26, y, 20, 1)

	d.to.Resize(x+60, y, 20, 1)
}

func (d *DateFilterPanel) Render(updateScreen bool) {
	if !d.IsVisible() {
		return
	}

	d.ColoredPanel.Render(false)

	style := d.ColoredPanel.CurrentStyler.Style()

	header := fmt.Sprintf(" %s", d.Name())
	_, y := d.Position()
	components.RenderText(0, y, header, style.Reverse(true))
	x := components.RenderText(19, y, "From ", style.Reverse(true))
	components.RenderText(x, y, "▶ ", style)
	x = components.RenderText(55, y, "To ", style.Reverse(true))
	components.RenderText(x, y, "▶ ", style)

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
	d.ColoredPanel.SetColorIndex(colorIndex)

	d.from.SetColorIndex(colorIndex)
	d.to.SetColorIndex(colorIndex)

	if d.Filter() != nil {
		model.GetFilterManager().UpdateFilterColorIndex(d.Filter(), colorIndex)
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
	d.ColoredPanel.SetActive(active)

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
	d.ColoredPanel.SetFilter(filter)

	d.from.SetFilter(filter)
	d.to.SetFilter(filter)
}
