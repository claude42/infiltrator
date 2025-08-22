package components

import (
	"fmt"

	"github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

type Select struct {
	ComponentImpl
	util.ObservableImpl

	Options  []string
	selected int

	OldStyler     Styler
	CurrentStyler Styler
}

func NewSelect(options []string) *Select {
	s := &Select{}
	s.Options = options
	// just to set the initial width correctly, mb not even necessary?!
	s.Resize(-1, -1, -1, -1)
	s.StyleUsing(s)

	return s
}

func (s *Select) Position() (int, int) {
	return s.x, s.y
}

func (s *Select) Width() int {
	return s.width
}

func (s *Select) SetOptions(options []string) {
	s.Options = options
}

func (s *Select) SetSelectedIndex(selected int) error {
	if selected >= len(s.Options) {
		return util.ErrOutOfBounds
	}
	s.selected = selected
	return nil
}

func (s *Select) SelectedIndex() int {
	return s.selected
}

func (s *Select) SelectedOption() string {
	return s.Options[s.selected]
}

func (s *Select) updateWidth() (width int) {
	width = 0
	for _, option := range s.Options {
		width = max(len(option), width)
	}
	return
}

func (s *Select) Resize(x, y, width, height int) {
	// width and height get ignored
	s.ComponentImpl.Resize(x, y, s.updateWidth(), 1)
}

func (s *Select) Height() int {
	return 1
}

func (s *Select) Size() (int, int) {
	return s.width, 1
}

func (s *Select) Render(updateScreen bool) {
	if !s.visible {
		return
	}

	str := fmt.Sprintf("[%-*s]", s.width, s.Options[s.selected])
	RenderText(s.x, s.y, str, s.CurrentStyler.Style())
}

func (s *Select) Style() tcell.Style {
	style := tcell.StyleDefault.Reverse(true)

	if s.IsActive() {
		return style.Bold(true)
	} else {
		return style
	}
}

func (s *Select) NextOption() int {
	s.selected++
	if s.selected >= len(s.Options) {
		s.selected = 0
	}
	return s.selected
}

func (s *Select) StyleUsing(styler Styler) {
	if s.CurrentStyler != nil {
		s.OldStyler = s.CurrentStyler
	}
	s.CurrentStyler = styler
}
