package ui

import (
	"log"
	"strings"

	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

type Orientation int

const (
	OrientationLeft Orientation = iota
	OrientationCenter
	OrientationRight
)

type Modal interface {
	Component
	SetTitle(title string)
}

type ModalImpl struct {
	ComponentImpl

	x, y, width, height int
	title               string
	orientation         Orientation
	lines               []string
}

func NewModalImpl(width, height int) *ModalImpl {
	m := &ModalImpl{}
	m.Resize(0, 0, width, height)

	return m
}

func NewModalImplWithContent(content string, orientation Orientation) *ModalImpl {
	m := &ModalImpl{}
	m.SetContent(content, orientation)

	return m
}

func (m *ModalImpl) SetTitle(title string) {
	m.title = title
}

func (m *ModalImpl) HandleEvent(ev tcell.Event) bool {
	switch ev.(type) {
	case *model.EventDisplay:
		m.Render(true)
		return false
	}

	return false
}

// special case: Resize(0, 0, 0, 0): don't resize, just adjust to screen if
// necessary
func (m *ModalImpl) Resize(x, y, width, height int) {
	// x and y are ignore - modal will always be centered
	// TODO: change Resize() interface so it can return an error
	screenWidth, screenHeight := screen.Size()

	if width != 0 {
		m.width = width
	}
	if height != 0 {
		m.height = height
	}

	m.x = util.IntMax((screenWidth-m.width)/2, 0)
	m.y = util.IntMax((screenHeight-m.height)/2, 0)
}

func (m *ModalImpl) Render(updateScreen bool) {
	if !m.active {
		return
	}

	for x := m.x; x < m.x+m.width; x++ {
		for y := m.y; y < m.y+m.height; y++ {
			screen.SetContent(x, y, ' ', nil, ModalStyle)
		}
	}

	m.renderTitle()

	if len(m.lines) > 0 {
		m.renderTextContent()
	}

	if updateScreen {
		screen.Show()
	}
}

func (m *ModalImpl) renderTitle() {
	centerText(m.x, m.width, m.y, m.title, ModalStyle)
}

func (m *ModalImpl) renderTextContent() {
	for i, line := range m.lines {
		switch m.orientation {
		case OrientationCenter:
			centerText(m.x, m.width+2, m.y+2+i, line, ModalStyle)
		case OrientationRight:
			log.Panicf("Orientation %d not implemented", m.orientation)
		case OrientationLeft:
			fallthrough
		default:
			renderText(m.x+2, m.y+2+i, line, ModalStyle)
		}
	}
}

func (m *ModalImpl) SetContent(content string, orientation Orientation) {
	m.lines = strings.Split(content, "\n")
	var contentWidth int
	for _, line := range m.lines {
		if len(line) > contentWidth {
			contentWidth = len(line)
		}
	}

	m.orientation = orientation

	m.Resize(0, 0, contentWidth+4, len(m.lines)+3)
}
