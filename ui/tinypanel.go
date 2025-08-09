package ui

import (
	//"fmt"
	"fmt"
	"log"

	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"

	//"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

const tinyPanelDefaultName = "This should not be seen here"

const nameWidth = 9
const headerWidth = 24

var filterModes = []string{
	"focus",
	"match",
	"hide",
}

var caseSensitive = []string{
	"case",
	"CaSe",
}

type TinyPanel struct {
	name          string
	y             int
	width         int
	input         *FilterInput
	mode          *Select
	caseSensitive *Select
	colorIndex    uint8
	filter        filter.Filter

	ComponentImpl
}

func NewTinyPanel() *TinyPanel {
	t := &TinyPanel{name: tinyPanelDefaultName}
	t.input = NewFilterInput()
	t.mode = NewSelect(filterModes)
	t.caseSensitive = NewSelect(caseSensitive)

	return t
}

func (t *TinyPanel) Height() int {
	return 1
}

func (t *TinyPanel) Position() (int, int) {
	return 0, t.y
}

func (t *TinyPanel) Resize(x, y, width, height int) {
	// x, height get ignored
	t.y = y
	t.width = width

	t.input.Resize(x+headerWidth+2, y, width-(x+headerWidth+2), 1)
	t.mode.Resize(x+nameWidth, y, 1, 1)
	t.caseSensitive.Resize(x+nameWidth+8, y, 1, 1)
}

func (t *TinyPanel) Render(updateScreen bool) {
	style := t.determinePanelStyle()

	header := fmt.Sprintf(" %s", t.name)
	x := renderText(0, t.y, header, style.Reverse(true))
	drawChars(x, t.y, headerWidth-x, ' ', style.Reverse((true)))
	renderText(headerWidth, t.y, "► ", style)

	if t.input != nil {
		t.input.Render(updateScreen)
	}

	if t.mode != nil {
		t.mode.Render(updateScreen)
	}

	if t.caseSensitive != nil {
		t.caseSensitive.Render(updateScreen)
	}

	if updateScreen {
		screen.Show()
	}
}

func (t *TinyPanel) SetColorIndex(colorIndex uint8) {
	t.colorIndex = colorIndex
	t.input.SetColorIndex(colorIndex)
	t.mode.SetColorIndex(colorIndex)
	t.caseSensitive.SetColorIndex(colorIndex)
	if t.filter != nil {
		model.GetFilterManager().UpdateFilterColorIndex(t.filter, colorIndex)
	}
}

func (t *TinyPanel) determinePanelStyle() tcell.Style {
	if t.IsActive() {
		return tcell.StyleDefault.Bold(true).Foreground(FilterColors[t.colorIndex][0])
	} else {
		return tcell.StyleDefault.Foreground(FilterColors[t.colorIndex][1])
	}
}

func (t *TinyPanel) SetContent(content string) {
	if t.input == nil {
		log.Panicln("TinyPanel.SetContent() called without input field!")
		return
	}
	t.input.SetContent(content)

}

func (t *TinyPanel) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlS:
			t.toggleMode()
			return true
		case tcell.KeyCtrlH:
			t.toggleCaseSensitive()
			return true
		}
	case *tcell.EventMouse:
		buttons := ev.Buttons()
		if buttons&tcell.ButtonPrimary != 0 {
			if t.mouseToggleMode(ev) {
				return true
			}
			if t.mouseToggleCaseSensitive(ev) {
				return true
			}

		}
	}

	if t.input != nil && t.input.HandleEvent(ev) {
		return true
	}

	return false
}

func (t *TinyPanel) SetActive(active bool) {
	t.ComponentImpl.SetActive(active)
	t.input.SetActive(active)
	t.mode.SetActive(active)
	t.caseSensitive.SetActive(active)
}

func (t *TinyPanel) SetName(name string) {
	t.name = name
}

func (t *TinyPanel) SetFilter(filter filter.Filter) {
	t.filter = filter
	t.input.SetFilter(filter)
}

func (t *TinyPanel) Filter() filter.Filter {
	return t.filter
}

func (t *TinyPanel) WatchInput(eh tcell.EventHandler) {
	if t.input == nil {
		log.Panicln("TinyPanel.WatchInput() called without input field!")
		return
	}
	t.input.Watch(eh)
}

func (t *TinyPanel) toggleMode() {
	model.GetFilterManager().UpdateFilterMode(t.filter, filter.FilterMode(t.mode.NextOption()))

	t.Render(true)
}

func (t *TinyPanel) mouseToggleMode(ev *tcell.EventMouse) bool {
	mouseX, mouseY := ev.Position()
	if mouseX >= t.mode.x && mouseX <= t.mode.x+t.mode.width &&
		mouseY == t.mode.y {
		t.toggleMode()
		return true
	} else {
		return false
	}
}

func (t *TinyPanel) toggleCaseSensitive() {
	model.GetFilterManager().UpdateFilterCaseSensitiveUpdate(t.filter, t.caseSensitive.NextOption() != 0)

	t.Render(true)
}

func (t *TinyPanel) mouseToggleCaseSensitive(ev *tcell.EventMouse) bool {
	mouseX, mouseY := ev.Position()
	if mouseX >= t.caseSensitive.x &&
		mouseX <= t.caseSensitive.x+t.caseSensitive.width &&
		mouseY == t.caseSensitive.y {
		t.toggleCaseSensitive()
		return true
	} else {
		return false
	}
}

func (t *TinyPanel) Mode() filter.FilterMode {
	return filter.FilterMode(t.mode.SelectedIndex())
}

func (t *TinyPanel) SetMode(mode int) {
	t.mode.SetSelectedIndex(mode)
}
