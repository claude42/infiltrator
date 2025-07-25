package ui

import (
	"fmt"
	"log"

	"github.com/claude42/infiltrator/model"

	"github.com/gdamore/tcell/v2"
)

var window *Window
var screen tcell.Screen

type Window struct {
	mainView     *View
	BottomPanels []Panel
	PanelsOpen   bool
	activePanel  Panel
}

func Setup(pipeline *model.Pipeline) *Window {
	if window != nil {
		log.Panicln("ui.setup() called twice!")
	}

	window = &Window{}

	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Panicf("%+v", err)
	}

	if err = screen.Init(); err != nil {
		log.Panicf("%+v", err)
	}

	window.SetView(NewView(pipeline))

	var panel Panel
	panel, err = NewPanel(TypeRegex, model.FilterHighlight)
	if err != nil {
		log.Panicf("%+v", err)
	}

	window.AddPanel(panel)
	window.PanelsOpen = true

	setupScreen()

	window.resize()

	return window
}

func setupScreen() {
	screen.SetStyle(DefStyle)
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()
}

func Cleanup() {
	screen.Fini()
	window = nil
}

func (w *Window) Render() {
	w.mainView.Render(false)

	for _, p := range w.BottomPanels {
		p.Render(false)
	}

	screen.Show()
}

func (w *Window) EventLoop(quit chan<- struct{}) {
	for {
		ev := screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
			case tcell.KeyBacktab:
				err := w.switchPanel(-1)
				if err != nil {
					screen.Beep()
				}
				continue
			case tcell.KeyTab:
				err := w.switchPanel(1)
				if err != nil {
					screen.Beep()
				}
				continue
			case tcell.KeyCtrlP:
				panel, err := NewPanel(TypeKeyword, model.FilterHighlight)
				if err != nil {
					log.Panicf("%+v", err)
				}
				w.AddPanel(panel)
				if err != nil {
					log.Panicf("%+v", err)
					screen.Beep()
				}
				continue
			case tcell.KeyCtrlO:
				toBeDestroyed := w.activePanel
				err := w.RemovePanel()
				if err != nil {
					screen.Beep()
				}
				DestroyPanel(toBeDestroyed)
				w.Render()
				continue
			case tcell.KeyEscape, tcell.KeyCtrlC:
				close(quit)
				return
			case tcell.KeyCtrlL:
				w.resizeAndRedraw()
				continue
			}
		case *tcell.EventResize:
			w.resizeAndRedraw()
			// TODO: maybe change this in the future and let it trickle down
			// instead of calling resize() manually
			continue
		}

		if w.mainView.HandleEvent(ev) {
			continue
		}

		if w.activePanel.HandleEvent(ev) {
			continue
		}
	}
}

func (w *Window) resizeAndRedraw() {
	w.resize()
	w.Render()
	screen.Sync()
}

// Good to know: Render() is never called from resize().
// resize() will only every be initiated when resizing the shell (i.e. through
// an event) and not from a panel resize. And in these case window.Render()
// will be immediately called after window.resize().

func (w *Window) resize() {
	width, height := screen.Size()

	totalPanelHeight := w.totalPanelHeight()
	if totalPanelHeight >= height {
		w.PanelsOpen = false
		w.mainView.Resize(0, 0, width, height)
	} else {
		w.mainView.Resize(0, 0, width, height-totalPanelHeight)
		y := height - totalPanelHeight
		for _, p := range w.BottomPanels {
			p.Resize(0, y, width, 0) // x and height ignored
			y += p.Height()
		}
	}
}

func (w *Window) totalPanelHeight() int {
	if !w.PanelsOpen {
		return 0
	}

	var totalPanelHeight = 0

	for _, p := range w.BottomPanels {
		totalPanelHeight += p.Height()
	}

	return totalPanelHeight
}

func (w *Window) AddPanel(newPanel Panel) error {
	// TODO: return error if total height of panels would exceed screen height
	w.BottomPanels = append(w.BottomPanels, newPanel)
	w.SetActivePanel(newPanel)

	// resize() doesn't sound right here but will actually recalculate where
	// the panels should be placed and how big they are.
	w.resize()
	w.Render()
	return nil
}

func (w *Window) RemovePanel() error {
	if len(w.BottomPanels) == 1 {
		return fmt.Errorf("cannot remove last panel")
	}

	var newActivePanel Panel
	activePanelIndex := w.activePanelIndex()

	if activePanelIndex > 0 {
		newActivePanel = w.BottomPanels[activePanelIndex-1]
	} else {
		newActivePanel = w.BottomPanels[1]
	}

	w.BottomPanels = append(w.BottomPanels[:activePanelIndex],
		w.BottomPanels[activePanelIndex+1:]...)
	w.SetActivePanel(newActivePanel)

	w.resize()
	w.Render()

	return nil
}

func (w *Window) SetActivePanel(p Panel) {
	if w.activePanel != nil {
		w.activePanel.SetActive(false)
	}
	w.activePanel = p
	w.activePanel.SetActive(true)

	// is there any case where the whole window (instead of the affected panel)
	// would have to be redrawn?
	// w.Render()
}

func (w *Window) activePanelIndex() int {
	for i, panel := range w.BottomPanels {
		if panel == w.activePanel {
			return i
		}
	}
	log.Panicln("Panel not found")
	return -1 // never reached
}

func (w *Window) switchPanel(offset int) error {
	newPanelIndex := w.activePanelIndex() + offset

	if newPanelIndex < 0 || newPanelIndex >= len(w.BottomPanels) {
		return fmt.Errorf("no panel at index %d", newPanelIndex)
	}

	w.SetActivePanel(w.BottomPanels[newPanelIndex])
	// It would probably be more natural to call render within the SetActivePanel()
	// (or even the individual SetActive() methods of the panels and InputFields),
	// but this way we avoid unnecessary redraws when switching panels.
	w.Render()

	return nil
}

func (w *Window) SetView(v *View) {
	window.mainView = v
}

// As soon as we get more options, we should use a struct for this
func (w *Window) ShowLineNumbers(showLineNumbers bool) {
	w.mainView.SetShowLineNumbers(showLineNumbers)
}
