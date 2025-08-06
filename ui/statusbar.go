package ui

import (
	"fmt"
	"sync"

	// "log"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"

	// "github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

type StatusBarState int

const (
	StatusDefault StatusBarState = iota
	SatusFollow
	StatusPanelsOpen
	StatusHelp
)

const StatusDefaultText = "[/] search [n/N] next/previous match [F] follow"
const StatusFollow = StatusDefault
const StatusPanelOpenText = "[CTRL-S] change mode [CTRL-H] change case sensitive [CTRL-P/O] add remove panel"

type Statusbar struct {
	ComponentImpl
	sync.Mutex

	y                      int
	width                  int
	height                 int
	colorIndex             uint8
	percentage             int
	panelsOpen             bool
	busyVisualizationIndex int
	busyState              model.BusyState
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
	s.Mutex.Lock()
	drawChars(0, s.y, s.width, ' ', StatusBarStyle)

	if s.panelsOpen {
		s.renderPanelOpenStatusBar()
	} else if config.GetConfiguration().FollowFile {
		s.renderFollowStausBar()
	} else {
		s.renderDefaultStatusBar()
	}

	s.renderBusyVisualization()
	s.Mutex.Unlock()

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

func (s *Statusbar) renderBusyVisualization() {
	var toRender rune
	if s.busyState == model.Idle {
		toRender = ' '
	} else {
		toRender = s.bumpBusyState()
	}

	screen.SetContent(s.width-1, s.y, toRender, nil, StatusBarStyle)
}

func (s *Statusbar) bumpBusyState() rune {
	var busyVisualization = []rune{'|', '/', '-', '\\', '|', '/', '-', '\\'}

	s.busyVisualizationIndex++
	if s.busyVisualizationIndex >= len(busyVisualization) {
		s.busyVisualizationIndex = 0
	}

	return busyVisualization[s.busyVisualizationIndex]
}

func (s *Statusbar) SetColorIndex(colorIndex uint8) {
	s.colorIndex = colorIndex
}

func (s *Statusbar) Height() int {
	return s.height
}

func (s *Statusbar) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *model.EventBusySpinnerUpdate:
		s.busyState = ev.BusyState
		s.renderBusyVisualization()
		screen.Show()
	case *model.EventDisplay:
		s.percentage = ev.Display.Percentage
		s.renderPercentage()
		screen.Show()
	case *model.EventFileChanged:
		// don't call Render() here because otherwise the spinner would get
		// updaten way to frequently.
		// But the question is: shouldn't we not also limit the update rate
		// of the percentage as well?
		s.percentage = ev.Percentage()
		s.renderPercentage()
		screen.Show()
	case *EventPanelStateChanged:
		s.panelsOpen = ev.PanelsOpen()
		s.Render(true)
	}

	return false
}
