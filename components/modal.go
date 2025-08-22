package components

import (
	"log"
	"strings"

	"github.com/claude42/infiltrator/model"
	"github.com/gdamore/tcell/v2"
)

type Orientation int

const (
	OrientationLeft Orientation = iota
	OrientationCenter
	OrientationRight
)

type Modal interface {
	Container
	SetTitle(title string)
}

type ModalImpl struct {
	ContainerImpl

	title        string
	orientation  Orientation
	lines        []string
	contentWidth int
}

func NewModalImpl(width, height int) *ModalImpl {
	m := &ModalImpl{}
	m.Resize(-1, -1, width, height)

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
	// if !m.active {
	// 	return false
	// }

	switch ev.(type) {
	case *model.EventDisplay:
		m.Render(true)
		return false
	}

	return false
}

func (m *ModalImpl) Resize(x, y, width, height int) {
	// x and y are ignore - modal will always be centered
	// TODO: change Resize() interface so it can return an error
	screenWidth, screenHeight := Screen.Size()

	if width == -1 {
		width = m.width
	}

	if height == -1 {
		height = m.height
	}

	m.ContainerImpl.Resize(max((screenWidth-width)/2, 0),
		max((screenHeight-height)/2, 0), width, height)
}

func (m *ModalImpl) Render(updateScreen bool) {
	if !m.visible {
		return
	}

	m.ContainerImpl.Render(false)

	for x := m.x; x < m.x+m.width; x++ {
		for y := m.y; y < m.y+m.height; y++ {
			Screen.SetContent(x, y, ' ', nil, ModalStyle)
		}
	}

	m.renderTitle()

	if len(m.lines) > 0 {
		m.renderTextContent()
	}

	if updateScreen {
		Screen.Show()
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
			RenderText(m.x+2, m.y+2+i, line, ModalStyle)
		}
	}
}

func (m *ModalImpl) SetContent(content string, orientation Orientation) {
	m.lines = strings.Split(content, "\n")

	m.contentWidth = 0
	for _, line := range m.lines {
		if len(line) > m.contentWidth {
			m.contentWidth = len(line)
		}
	}

	m.orientation = orientation

	m.Fit()
}

func (m *ModalImpl) Fit() {
	if len(m.lines) == 0 {
		return
	}

	m.Resize(-1, -1, m.contentWidth+4, len(m.lines)+3)
}
