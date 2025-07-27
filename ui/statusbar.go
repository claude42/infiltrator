package ui

import (
	"fmt"
	"log"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"

	// "github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

const (
	StatusDefault int = iota
	StatusSearch
	StatusHelp
)

const StatusDefaultText = "[/] search [n/N] next/previous match [F] follow"
const StatusPanelOpenText = "[CTRL-S] change mode [CTRL-H] change case sensitive [CTRL-P/O] add remove panel"

type Statusbar struct {
	ComponentImpl

	y             int
	width         int
	height        int
	colorIndex    uint8
	currentStatus int
}

func NewStatusbar() *Statusbar {
	s := &Statusbar{}
	s.height = 1
	model.GetPipeline().Watch(s)

	return s
}

func (s *Statusbar) Resize(x, y, width, height int) {
	s.y = y
	s.width = width
}

func (s *Statusbar) Render(updateScreen bool) {
	drawChars(0, s.y, s.width, ' ', StatusBarStyle)

	switch s.currentStatus {
	case StatusDefault:
		s.renderPanelOpenStatusBar()
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

func (s *Statusbar) renderPanelOpenStatusBar() {
	s.renderPercentage()

	s.renderFileName()

	renderText(0, s.y, StatusPanelOpenText, StatusBarStyle)
}

func (s *Statusbar) renderFileName() {
	fileNameStr := fmt.Sprintf("\"%s\"", config.GetConfiguration().FileName)
	length := len(fileNameStr)
	start := s.width - length - 4 - 5

	renderText(start, s.y, fileNameStr, StatusBarStyle)
}

func (s *Statusbar) renderPercentage() {
	percentage, err := model.GetPipeline().Percentage()

	var percentStr string
	if err != nil {
		percentStr = ""
	} else {
		percentStr = fmt.Sprintf("%3d%%", percentage)
	}

	renderText(s.width-5, s.y, percentStr, StatusBarStyle)
}

func (s *Statusbar) SetColorIndex(colorIndex uint8) {
	s.colorIndex = colorIndex
}

func (s *Statusbar) Height() int {
	return s.height
}

func (s *Statusbar) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *model.EventFileChanged, *model.EventFilterOutput, *model.EventPositionChanged:
		log.Printf("Event: %T, %+v", ev, ev)
		s.Render(true)
	}

	return false
}
