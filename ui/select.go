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
	style       tcell.Style

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
	if selected >= len(options) {
		return util.ErrOutOfBounds
	}
	s.selected = selected
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

func (s *Select) SetStyle(style tcell.Style) {
	s.style = style
}

func (s *Select) Render(updateScreen bool) {
	str := fmt.Sprintf("[%-*s]", s.width, s.options[s.selected])
	renderText(s.x, s.y, str, s.style)
}
