package ui

import (
	"fmt"
	"sync"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/busy"
	"github.com/claude42/infiltrator/util"

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
	components.ComponentImpl
	sync.Mutex

	colorIndex             uint8
	percentage             int
	panelsOpen             bool
	busyVisualizationIndex int
	busyState              busy.State
}

func NewStatusbar() *Statusbar {
	s := &Statusbar{}
	model.GetFilterManager().Watch(s)

	return s
}

func (s *Statusbar) Resize(x, y, width, height int) {
	s.ComponentImpl.Resize(0, y, width, 1)
}

func (s *Statusbar) Height() int {
	return 1
}

func (s *Statusbar) Size() (int, int) {
	return s.Width(), 1
}

func (s *Statusbar) Render(updateScreen bool) {
	if !s.IsVisible() {
		return
	}

	_, y := s.ComponentImpl.Position()

	s.Mutex.Lock()
	components.DrawChars(0, y, s.Width(), ' ', StatusBarStyle)

	if s.panelsOpen {
		s.renderPanelOpenStatusBar()
	} else if config.UserCfg().Follow {
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
	// components.RenderText(s.x, s.y, str, s.determineStyle())
}

func (s *Statusbar) renderDefaultStatusBar() {
	s.renderPercentage()

	s.renderFileName()

	s.renderStatusDefaultText()
}

func (s *Statusbar) renderFollowStausBar() {
	s.renderFollow()

	s.renderFileName()

	s.renderStatusDefaultText()
}

func (s *Statusbar) renderStatusDefaultText() {
	_, y := s.ComponentImpl.Position()
	components.RenderText(0, y, StatusDefaultText, StatusBarStyle)
}

func (s *Statusbar) renderPanelOpenStatusBar() {
	s.renderPercentage()

	s.renderFileName()

	_, y := s.ComponentImpl.Position()
	components.RenderText(0, y, StatusPanelOpenText, StatusBarStyle)
}

func (s *Statusbar) renderFileName() {
	const spacer = 4
	const percentLength = 9
	fileNameStr := fmt.Sprintf("\"%s\"", config.UserCfg().FileName)
	length := len(fileNameStr)
	start := s.Width() - length - spacer - percentLength

	_, y := s.ComponentImpl.Position()
	components.RenderText(start, y, fileNameStr, StatusBarStyle)
}

func (s *Statusbar) renderPercentage() {
	var percentStr string

	realPercentage, _ := util.InBetween(s.percentage, 0, 100)
	percentStr = fmt.Sprintf("%3d%%", realPercentage)

	var style tcell.Style

	if s.busyState != busy.Busy {
		style = StatusBarStyle
	} else {
		style = StatusBarBusyStyle
	}

	_, y := s.ComponentImpl.Position()
	components.RenderText(s.Width()-5, y, percentStr, style)
}

func (s *Statusbar) renderFollow() {
	_, y := s.ComponentImpl.Position()
	components.RenderText(s.Width()-9, y, "[follow]", StatusBarStyle)
}

func (s *Statusbar) renderBusyVisualization() {
	var toRender rune
	var style tcell.Style
	if s.busyState == busy.Idle {
		toRender = ' '
		style = StatusBarStyle
	} else {
		toRender = s.bumpBusyVisualization()
		style = StatusBarBusyStyle
	}

	_, y := s.ComponentImpl.Position()
	screen.SetContent(s.Width()-1, y, toRender, nil, style)
}

func (s *Statusbar) bumpBusyVisualization() rune {
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

func (s *Statusbar) HandleEvent(ev tcell.Event) bool {
	// if !s.IsActive() {
	// 	return false
	// }

	switch ev := ev.(type) {
	case *busy.EventBusySpinnerUpdate:
		s.busyState = ev.BusyState
		if ev.BusyPercentage != -1 {
			s.percentage = int(ev.BusyPercentage)
		}
		s.renderBusyVisualization()
		s.renderPercentage()
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
