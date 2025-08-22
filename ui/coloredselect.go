package ui

import (
	"github.com/claude42/infiltrator/components"
	"github.com/gdamore/tcell/v2"
)

type ColoredSelect struct {
	components.Select

	colorIndex uint8
}

func NewColoredSelect(options []string, key tcell.Key, do func(int)) *ColoredSelect {
	s := &ColoredSelect{
		Select: *components.NewSelect(options, key, do),
	}

	s.StyleUsing(s)

	return s
}

func (s *ColoredSelect) SetColorIndex(colorIndex uint8) {
	s.colorIndex = colorIndex
}

func (s *ColoredSelect) Style() tcell.Style {
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
