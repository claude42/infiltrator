package ui

import (
	"github.com/claude42/infiltrator/components"

	"github.com/gdamore/tcell/v2"
)

type ColoredInput struct {
	components.InputImpl

	colorIndex uint8
}

func NewColoredInput() *ColoredInput {
	c := &ColoredInput{
		InputImpl: *components.NewInputImpl(),
	}

	c.InputImpl.StyleUsing(c)

	return c
}

func (c *ColoredInput) SetColorIndex(colorIndex uint8) {
	c.colorIndex = colorIndex
}

func (c *ColoredInput) Style() tcell.Style {
	var style tcell.Style

	if c.InputImpl.OldStyler != nil {
		style = c.InputImpl.OldStyler.Style()
	} else {
		style = tcell.StyleDefault
	}

	if c.IsActive() {
		style = style.Foreground((FilterColors[c.colorIndex][0]))
	} else {
		style = style.Foreground((FilterColors[c.colorIndex][1]))
	}

	if !c.InputImpl.InputCorrect {
		style = style.Italic(true)
	}

	return style
}
