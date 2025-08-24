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
					GetPanelManager().CreateAndAdd(config.FilterTypeRegex)
					GetPanelManager().SetPanelsOpen(true)
					p.Hide()
					components.RenderAll(true)
					return true
				case 'k':
					GetPanelManager().CreateAndAdd(config.FilterTypeKeyword)
					GetPanelManager().SetPanelsOpen(true)
					p.Hide()
					components.RenderAll(true)
					return true
				case 'd':
					GetPanelManager().CreateAndAdd(config.FilterTypeDate)
					GetPanelManager().SetPanelsOpen(true)
					p.Hide()
					components.RenderAll(true)
					return true
				}
				return true
			case tcell.KeyEscape:
				p.Hide()
				components.RenderAll(true)
				return true
			case tcell.KeyEnter:
				GetPanelManager().CreateAndAdd(config.FilterTypeRegex)
				GetPanelManager().SetPanelsOpen(true)
				p.Hide()
				components.RenderAll(true)
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

	GetScreen().PostEvent(NewEventPopupStateChanged(popupState, p))
}
