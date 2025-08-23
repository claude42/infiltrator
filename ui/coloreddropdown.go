package ui

import (
	"github.com/claude42/infiltrator/components"
	"github.com/gdamore/tcell/v2"
)

type ColoredDropdown struct {
	components.Dropdown

	colorIndex uint8
}

func NewColoredDropdown(options []string, key tcell.Key, do func(int)) *ColoredDropdown {
	s := &ColoredDropdown{
		Dropdown: *components.NewDropdown(options, key, do),
	}

	s.StyleUsing(s)

	return s
}

func (s *ColoredDropdown) SetColorIndex(colorIndex uint8) {
	s.colorIndex = colorIndex
}

func (s *ColoredDropdown) Style() tcell.Style {
	var style tcell.Style
	if s.OldStyler != nil {
		style = s.OldStyler.Style()
	} else {
		style = tcell.StyleDefault.Reverse(true)
	}

	if s.IsActive() {
		style = style.Foreground((FilterColors[s.colorIndex][0]))
	} else {
		style = style.Foreground((FilterColors[s.colorIndex][1]))
	}

	return style
}
