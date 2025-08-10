package ui

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

const content = `[ R ] Regular expression
[ K ] Simple keyword search
[ G ] Glob style pattern matching
[ D ] Date filter`

type PanelSelection struct {
	ModalImpl
}

func NewPanelSelection() *PanelSelection {
	p := &PanelSelection{}
	p.SetTitle("Choose type of filter")
	p.SetContent(content, OrientationLeft)
	// p.ModalImpl.Resize(0, 0, 0, 0)

	return p
}

func (p *PanelSelection) HandleEvent(ev tcell.Event) bool {
	log.Println("dochdrin")
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'r':
				window.CreateAndAddPanel(PanelTypeRegex)
				window.SetPanelsOpen(true)
				p.SetActive(false)
			case 'k':
				window.CreateAndAddPanel(PanelTypeKeyword)
				window.SetPanelsOpen(true)
				p.SetActive(false)
			case 'd':
				window.CreateAndAddPanel(PanelTypeDate)
				window.SetPanelsOpen(true)
				p.SetActive(false)
			}
			return true
		case tcell.KeyEscape:
			p.SetActive(false)
			return true
		case tcell.KeyEnter:
			window.CreateAndAddPanel(PanelTypeRegex)
			window.SetPanelsOpen(true)
			p.SetActive(false)
			return true
		}
	}

	return p.ModalImpl.HandleEvent(ev)
}

func (p *PanelSelection) SetActive(active bool) {
	p.ModalImpl.SetActive(active)

	var popupState PopupState
	if active {
		popupState = PopupPanelSelection
	} else {
		popupState = PopupNone
	}

	GetScreen().PostEvent(NewEventPopupStateChanged(popupState))
}
