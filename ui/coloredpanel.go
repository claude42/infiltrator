package ui

import (
	"log"

	"github.com/claude42/infiltrator/components"
	"github.com/gdamore/tcell/v2"
)

type ColoredPanel struct {
	components.PanelImpl

	colorIndex uint8
}

func NewColoredPanel(name string) *ColoredPanel {
	c := &ColoredPanel{
		PanelImpl: *components.NewPanelImpl(name),
	}

	log.Printf("Styler: %T\n%+v\n%p", c, c, c)
	c.PanelImpl.StyleUsing(c)

	return c
}

func (c *ColoredPanel) SetColorIndex(colorIndex uint8) {
	log.Printf("Styler: %T\n%+v\n%p", c, c, c)
	c.colorIndex = colorIndex
}

func (c *ColoredPanel) Style() tcell.Style {
	log.Printf("Styler: %T\n%+v\n%p", c, c, c)
	var style tcell.Style
	if c.PanelImpl.OldStyler != nil {
		style = c.PanelImpl.OldStyler.Style()
	} else {
		style = tcell.StyleDefault
	}

	if c.IsActive() {
		return style.Foreground(FilterColors[c.colorIndex][0])
	} else {
		return style.Foreground(FilterColors[c.colorIndex][1])
	}
}
