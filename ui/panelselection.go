package ui

import (
	"log"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"

	"github.com/gdamore/tcell/v2"
)

const content = `[ R ] Regular expression
[ K ] Simple keyword search
[ G ] Glob style pattern matching
[ D ] Date filter`

type PanelSelection struct {
	components.ModalImpl
}

func NewPanelSelection() *PanelSelection {
	p := &PanelSelection{}
	p.SetTitle("Choose type of filter")
	p.SetContent(content, components.OrientationLeft)
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
				window.CreateAndAddPanel(config.FilterTypeRegex)
				window.SetPanelsOpen(true)
				p.SetActive(false)
			case 'k':
				window.CreateAndAddPanel(config.FilterTypeKeyword)
				window.SetPanelsOpen(true)
				p.SetActive(false)
			case 'd':
				window.CreateAndAddPanel(config.FilterTypeDate)
				window.SetPanelsOpen(true)
				p.SetActive(false)
			}
			return true
		case tcell.KeyEscape:
			p.SetActive(false)
			return true
		case tcell.KeyEnter:
			window.CreateAndAddPanel(config.FilterTypeRegex)
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
