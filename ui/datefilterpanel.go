package ui

import (
	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
	"github.com/gdamore/tcell/v2"
)

type DateFilterPanel struct {
	*FilterPanelImpl

	typeSelect *ColoredDropdown
	from       *FilterInput
	to         *FilterInput
	lastActive *FilterInput
}

func NewDateFilterPanel(panelType config.FilterType, name string) *DateFilterPanel {

	d := &DateFilterPanel{
		FilterPanelImpl: NewFilterPanelImpl(panelType, name),
		from:            NewFilterInput(filter.DateFilterFrom),
		to:              NewFilterInput(filter.DateFilterTo),
	}
	d.typeSelect = NewColoredDropdown(config.Filters.AllStrings(), tcell.KeyCtrlH, d.changePanelType)
	d.typeSelect.SetSelectedIndex(int(panelType))
	d.Add(d.typeSelect)
	d.Add(d.from)
	d.Add(d.to)
	d.to.SetActive(false)
	return d
}

func (d *DateFilterPanel) SetPanelConfig(panelConfig *config.PanelTable) {
	if panelConfig == nil {
		return
	}

	d.SetFrom(panelConfig.From)
	d.SetTo(panelConfig.To)

	// don't put this into FilterPanelImpl!
	d.SetColorIndex(panelConfig.ColorIndex)
}

func (d *DateFilterPanel) Resize(x, y, width, height int) {
	d.FilterPanelImpl.Resize(x, y, width, height)

	d.typeSelect.Resize(x+1, y, config.PanelNameWidth, 1)
	d.from.Resize(x+config.PanelHeaderWidth+config.PanelHeaderGap, y, 20, 1)
	d.to.Resize(x+60, y, 20, 1)
}

func (d *DateFilterPanel) Render(updateScreen bool) {
	if !d.IsVisible() {
		return
	}

	d.FilterPanelImpl.Render(false)

	style := d.CurrentStyler.Style()

	_, y := d.Position()
	x := components.RenderText(config.PanelHeaderWidth-len("From "), y, "From ", style.Reverse(true))
	components.RenderText(x, y, "▶ ", style)
	x = components.RenderText(55, y, "To ", style.Reverse(true))
	components.RenderText(x, y, "▶ ", style)

	if updateScreen {
		screen.Show()
	}

}

func (d *DateFilterPanel) SetColorIndex(colorIndex uint8) {
	d.FilterPanelImpl.SetColorIndex(colorIndex)

	if d.Filter() != nil {
		model.GetFilterManager().UpdateFilterColorIndex(d.Filter(), colorIndex)
	}
}

func (d *DateFilterPanel) HandleEvent(ev tcell.Event) bool {
	if !d.IsActive() {
		return false
	}

	if d.IsActive() {
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
	}

	return d.FilterPanelImpl.HandleEvent(ev)
}

func (d *DateFilterPanel) SetActive(active bool) {
	d.FilterPanelImpl.SetActive(active)

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
	d.FilterPanelImpl.SetFilter(filter)

	d.from.SetFilter(filter)
	d.to.SetFilter(filter)
}

// TODO: test if these really work
func (d *DateFilterPanel) SetFrom(from string) {
	d.from.SetContent(from)
}

func (d *DateFilterPanel) From() string {
	return d.from.Content()
}

func (d *DateFilterPanel) SetTo(to string) {
	d.to.SetContent(to)
}

func (d *DateFilterPanel) To() string {
	return d.to.Content()
}

func (d *DateFilterPanel) changePanelType(i int) {
	newType := config.FilterType(i)
	if newType == d.panelType {
		return
	}

	d.panelConfig.From = d.From()
	d.panelConfig.To = d.To()
	d.panelConfig.ColorIndex = d.ColorIndex()

	// Note-to-self: don't put the next lines into FilterPanelImpl!
	newPanel := NewPanelWithPanelTypeAndConfig(newType, &d.panelConfig)
	newPanel.Show()
	err := window.ReplacePanel(d, newPanel)
	fail.OnError(err, "failed to replace panel")
}
