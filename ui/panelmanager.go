package ui

import (
	"fmt"
	"log"
	"sync"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
)

type PanelManager struct {
	panels      []FilterPanel
	panelsOpen  bool
	activePanel FilterPanel
}

var (
	pmInstance *PanelManager
	pmOnce     sync.Once
)

func GetPanelManager() *PanelManager {
	pmOnce.Do(func() {
		pmInstance = &PanelManager{
			panels: make([]FilterPanel, 0),
		}
	})
	return pmInstance
}

func (pm *PanelManager) openPanelsOrPanelSelection() {
	// if panels are currently closed but at least one panel exists
	// already, then just open the existing panels, don't open
	// panel selection
	if pm.activePanel != nil && !pm.panelsOpen {
		pm.SetPanelsOpen(true)
		return
	} else {
		// if w.popup != nil { // TODO: weird
		// 	w.popup.SetActive(false)
		// }
		window.popup = window.panelSelection
		window.popup.Show()
	}
}

func (pm *PanelManager) SetPanelsOpen(panelsOpen bool) {
	pm.panelsOpen = panelsOpen
	if panelsOpen {
		for _, p := range pm.panels {
			p.SetVisible(true)
		}
		pm.activePanel.SetActive(true)
	} else {
		for _, p := range pm.panels {
			p.Hide()
		}
	}
	GetScreen().PostEvent(NewEventPanelStateChanged(panelsOpen))
	window.resizeAndRedraw()
}

func (pm *PanelManager) PanelsOopen() bool {
	return pm.panelsOpen
}

func (pm *PanelManager) CreateAndAdd(panelType config.FilterType) {
	if pm.activePanel != nil && !pm.panelsOpen {
		return
	}

	pm.Add(NewPanel(panelType))
}

func (pm *PanelManager) Add(newPanel FilterPanel) error {
	if newPanel == nil {
		return nil
	}
	// TODO: return error if total height of panels would exceed screen height
	pm.panels = append(pm.panels, newPanel)
	components.Add(newPanel, 1)
	pm.SetActive(newPanel)

	// resize() doesn't sound right here but will actually recalculate where
	// the panels should be placed and how big they are.
	// Is this really necessary?
	// w.resize()
	// w.Render()
	return nil
}

func (pm *PanelManager) Remove() error {
	if len(pm.panels) == 1 {
		return fmt.Errorf("cannot remove last panel")
	}

	var newActivePanel FilterPanel
	activePanelIndex := pm.activePanelIndex()

	if activePanelIndex > 0 {
		newActivePanel = pm.panels[activePanelIndex-1]
	} else {
		newActivePanel = pm.panels[1]
	}

	pm.panels = append(pm.panels[:activePanelIndex],
		pm.panels[activePanelIndex+1:]...)
	components.Remove(pm.activePanel)
	DestroyPanel(pm.activePanel)
	pm.SetActive(newActivePanel)

	window.resize()
	window.Render()

	return nil
}

func (pm *PanelManager) Replace(oldPanel, newPanel FilterPanel) error {
	fail.If(oldPanel == nil || newPanel == nil, "old or new panel is nil")

	var found bool
	for i, p := range pm.panels {
		if p == oldPanel {
			pm.panels[i] = newPanel
			found = true
			break
		}
	}
	fail.If(!found, "old panel not found in window.BottomPanels")

	GetColorManager().Replace(oldPanel, newPanel) // must be called before DestroyPanel()
	components.Remove(oldPanel)
	DestroyPanel(oldPanel)
	components.Add(newPanel, 1)
	pm.SetActive(newPanel)
	window.resize()
	window.Render()

	return nil
}

func (pm *PanelManager) switchPanel(offset int) error {
	newPanelIndex := pm.activePanelIndex() + offset

	if newPanelIndex < 0 || newPanelIndex >= len(pm.panels) {
		return fmt.Errorf("no panel at index %d", newPanelIndex)
	}

	return pm.goTo(newPanelIndex)
}

func (pm *PanelManager) goTo(no int) error {
	if no < 0 || no >= len(pm.panels) {
		return fmt.Errorf("no panel at index %d", no)
	}

	pm.SetActive(GetPanelManager().panels[no])
	// It would probably be more natural to call render within the SetActivePanel()
	// (or even the individual SetActive() methods of the panels and InputFields),
	// but this way we avoid unnecessary redraws when switching panels.
	window.Render()

	return nil
}

func (pm *PanelManager) SetActive(p FilterPanel) {
	if pm.activePanel != nil {
		pm.activePanel.SetActive(false)
	}
	pm.activePanel = p
	pm.activePanel.SetActive(true)

	// is there any case where the whole window (instead of the affected panel)
	// would have to be redrawn?
	// w.Render()
}

func (pm *PanelManager) activePanelIndex() int {
	for i, panel := range pm.panels {
		if panel == pm.activePanel {
			return i
		}
	}
	log.Panicln("Panel not found")
	return -1 // never reached
}

func (pm *PanelManager) totalHeight() int {
	if !pm.panelsOpen {
		return 0
	}

	var totalHeight = 0

	for _, p := range pm.panels {
		totalHeight += p.Height()
	}

	return totalHeight
}

func (pm *PanelManager) copyPanelsToConfig() {
	// cfg := config.GetConfiguration()
	newPanels := make([]config.PanelTable, 0, len(pm.panels))

	for _, p := range pm.panels {
		var cp config.PanelTable
		switch p := p.(type) {
		case *StringFilterPanel:
			cp = config.PanelTable{
				Type:          p.Name(),
				Key:           p.Content(),
				Mode:          config.FilterModeStrings[p.Mode()],
				CaseSensitive: p.CaseSensitive(),
			}
		case *DateFilterPanel:
			cp = config.PanelTable{
				Type: p.Name(),
				From: p.From(),
				To:   p.To(),
			}
		default:
			log.Fatalf("Unknown panel type %s", p.Name())
		}

		newPanels = append(newPanels, cp)
	}
	config.SetPanels(newPanels)
}
