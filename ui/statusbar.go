package ui

import (
	"fmt"
	// "log"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"

	// "github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

const (
	StatusDefault int = iota
	SatusFollow
	StatusPanelsOpen
	StatusHelp
)

const StatusDefaultText = "[/] search [n/N] next/previous match [F] follow"
const StatusFollow = StatusDefault
const StatusPanelOpenText = "[CTRL-S] change mode [CTRL-H] change case sensitive [CTRL-P/O] add remove panel"

type Statusbar struct {
	ComponentImpl

	y             int
	width         int
	height        int
	colorIndex    uint8
	currentStatus int
	percentage    int
	panelsOpen    bool
}

func NewStatusbar() *Statusbar {
	s := &Statusbar{}
	s.height = 1
	model.GetFilterManager().Watch(s)

	return s
}

func (s *Statusbar) Resize(x, y, width, height int) {
	s.y = y
	s.width = width
}

func (s *Statusbar) Render(updateScreen bool) {
	drawChars(0, s.y, s.width, ' ', StatusBarStyle)

	if s.panelsOpen {
		s.renderPanelOpenStatusBar()
	} else if config.GetConfiguration().FollowFile {
		s.renderFollowStausBar()
	} else {
		s.renderDefaultStatusBar()
	}

	if updateScreen {
		screen.Show()
	}
	// str := fmt.Sprintf("[%-*s]", s.width, s.options[s.selected])
	// renderText(s.x, s.y, str, s.determineStyle())
}

func (s *Statusbar) renderDefaultStatusBar() {
	s.renderPercentage()

	s.renderFileName()

	renderText(0, s.y, StatusDefaultText, StatusBarStyle)
}

func (s *Statusbar) renderFollowStausBar() {
	s.renderFollow()

	s.renderFileName()

	renderText(0, s.y, StatusDefaultText, StatusBarStyle)
}
func (s *Statusbar) renderPanelOpenStatusBar() {
	s.renderPercentage()

	s.renderFileName()

	renderText(0, s.y, StatusPanelOpenText, StatusBarStyle)
}

func (s *Statusbar) renderFileName() {
	const spacer = 4
	const percentLength = 9
	fileNameStr := fmt.Sprintf("\"%s\"", config.GetConfiguration().FileName)
	length := len(fileNameStr)
	start := s.width - length - spacer - percentLength

	renderText(start, s.y, fileNameStr, StatusBarStyle)
}

func (s *Statusbar) renderPercentage() {
	var percentStr string

	if s.percentage >= 0 && s.percentage <= 100 {
		percentStr = fmt.Sprintf("%3d%%", s.percentage)
	} else {
		percentStr = ""
	}

	renderText(s.width-5, s.y, percentStr, StatusBarStyle)
}

func (s *Statusbar) renderFollow() {
	renderText(s.width-9, s.y, "[follow]", StatusBarStyle)
}

func (s *Statusbar) SetColorIndex(colorIndex uint8) {
	s.colorIndex = colorIndex
}

func (s *Statusbar) Height() int {
	return s.height
}

func (s *Statusbar) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *model.EventDisplay:
		s.percentage = ev.Display.Percentage
		s.Render(true)
	case *model.EventFileChanged:
		s.percentage = ev.Percentage()
		s.Render(true)
	case *EventPanelStateChanged:
		s.panelsOpen = ev.PanelsOpen()
		s.Render(true)
	}

	return false
}
