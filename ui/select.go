package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

type Select struct {
	x, y, width int
	options     []string
	selected    int
	colorIndex  uint8

	ComponentImpl
}

func NewSelect(options []string) *Select {
	s := &Select{}
	s.options = options
	s.updateWidth()

	return s
}

func (s *Select) SetOptions(options []string) {
	s.options = options
}

func (s *Select) SetSelectedIndex(selected int) error {
	if selected >= len(s.options) {
		return util.ErrOutOfBounds
	}
	s.selected = selected
	return nil
}

func (s *Select) SelectedIndex() int {
	return s.selected
}

func (s *Select) SelectedOption() string {
	return s.options[s.selected]
}

func (s *Select) updateWidth() {
	s.width = 0
	for _, option := range s.options {
		s.width = util.IntMax(len(option), s.width)
	}
}

func (s *Select) Resize(x, y, width, height int) {
	// width and height get ignored
	s.x = x
	s.y = y
}

func (s *Select) Render(updateScreen bool) {
	str := fmt.Sprintf("[%-*s]", s.width, s.options[s.selected])
	renderText(s.x, s.y, str, s.determineStyle())
}

func (s *Select) SetColorIndex(colorIndex uint8) {
	s.colorIndex = colorIndex
}

func (s *Select) determineStyle() tcell.Style {
	style := tcell.StyleDefault.Reverse(true)

	if s.IsActive() {
		style = style.Foreground((FilterColors[s.colorIndex][0])).Bold(true)
	} else {
		style = style.Foreground((FilterColors[s.colorIndex][1]))
	}

	return style
}

func (s *Select) NextOption() int {
	s.selected++
	if s.selected >= len(s.options) {
		s.selected = 0
	}
	return s.selected
}
