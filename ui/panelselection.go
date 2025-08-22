package ui

import (
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
	p := &PanelSelection{
		ModalImpl: *components.NewModalImplWithContent(content, components.OrientationLeft),
	}
	p.SetTitle("Choose type of filter")

	return p
}

func (p *PanelSelection) HandleEvent(ev tcell.Event) bool {
	if p.IsActive() {
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'r':
					window.CreateAndAddPanel(config.FilterTypeRegex)
					window.SetPanelsOpen(true)
					p.Hide()
					return true
				case 'k':
					window.CreateAndAddPanel(config.FilterTypeKeyword)
					window.SetPanelsOpen(true)
					p.Hide()
					return true
				case 'd':
					window.CreateAndAddPanel(config.FilterTypeDate)
					window.SetPanelsOpen(true)
					p.Hide()
					return true
				}
				return true
			case tcell.KeyEscape:
				p.Hide()
				return true
			case tcell.KeyEnter:
				window.CreateAndAddPanel(config.FilterTypeRegex)
				window.SetPanelsOpen(true)
				p.Hide()
				return true
			}
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
