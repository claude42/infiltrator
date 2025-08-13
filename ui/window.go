package ui

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

var (
	screen tcell.Screen
	once   sync.Once
	window *Window
)

type Window struct {
	mainView       *View
	BottomPanels   []components.Panel
	panelsOpen     bool
	activePanel    components.Panel
	statusbar      *Statusbar
	popup          components.Modal
	panelSelection *PanelSelection
	exPanel        *ExPanel
}

func GetScreen() tcell.Screen {
	once.Do(func() {
		// switch to alternate buffer
		fmt.Print("\x1b[?1049h")
		screen = fail.Must1(tcell.NewScreen())
	})
	return screen
}

func InfiltPostEvent(ev util.Event) error {
	return GetScreen().PostEvent(ev)
}

func Setup() *Window {
	components.Screen = GetScreen()

	fail.If(window != nil, "ui.setup() called twice!")

	window = &Window{}

	fail.Must0(screen.Init())

	window.mainView = NewView()
	components.Add(window.mainView, 0)
	window.mainView.SetVisible(true)

	window.statusbar = NewStatusbar()
	components.Add(window.statusbar, 0)
	window.statusbar.SetVisible(true)

	window.panelSelection = NewPanelSelection()
	components.Add(window.panelSelection, 2)

	window.exPanel = NewExPanel()
	components.Add(window.exPanel, 1)
	window.exPanel.SetContent("Hallo")
	// window.exPanel.SetActive(true)

	setupScreen()

	return window
}

func setupScreen() {
	screen.SetStyle(DefStyle)
	screen.EnableMouse(tcell.MouseButtonEvents)
	screen.DisablePaste()
	screen.Clear()
}

func Cleanup() {
	screen.Fini()
	// switch back to primary buffer
	fmt.Print("\x1b[?1049l")
}

func (w *Window) Render() {
	components.RenderAll(false)

	screen.Show()
}

func (w *Window) MetaEventLoop(ctx context.Context, wg *sync.WaitGroup, quit chan<- string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("A panic occurred: %v\nStack trace:\n%s", r, debug.Stack())
			panic(r)
		}
	}()

	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("Received shutdown signal")
			return
		default:
			if w.EventLoop(quit) {
				return
			}
		}
	}
}

func (w *Window) EventLoop(quit chan<- string) bool {
	ev := screen.PollEvent()
	// log.Printf("Event: %T, %+v", ev, ev)
	log.Printf("Main Loop: %T", ev)

	if w.popup != nil && w.popup.IsActive() && w.popup.HandleEvent(ev) {
		return false
	}

	if w.panelsOpen && w.activePanel != nil &&
		w.activePanel.HandleEvent(ev) {

		return false
	}

	switch ev := ev.(type) {
	case *tcell.EventKey:
		// log.Print("-------------------------------------------------------")
		// log.Printf("Modmast: %d", ev.Modifiers())
		// log.Printf("Rune: %c", ev.Rune())
		// log.Printf("Key: %s", tcell.KeyNames[ev.Key()])
		log.Printf("Key Name: %s", ev.Name())

		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case '/':
				w.openPanelsOrPanelSelection()
				w.Render()
				return false
			case 'q':
				quit <- "Good bye!"
				close(quit)
				return true
			}
		case tcell.KeyBacktab:
			err := w.switchPanel(-1)
			if err != nil {
				screen.Beep()
			}
			return false
		case tcell.KeyTab:
			err := w.switchPanel(1)
			if err != nil {
				screen.Beep()
			}
			return false
		case tcell.KeyCtrlP:
			w.openPanelsOrPanelSelection()
			w.Render()
			return false
		case tcell.KeyCtrlO:
			toBeDestroyed := w.activePanel
			err := w.RemovePanel()
			if err != nil {
				screen.Beep()
			}
			DestroyPanel(toBeDestroyed)
			w.Render()
			return false
		case tcell.KeyF1, tcell.KeyF2, tcell.KeyF3, tcell.KeyF4, tcell.KeyF5, tcell.KeyF6,
			tcell.KeyF7, tcell.KeyF8, tcell.KeyF9, tcell.KeyF10, tcell.KeyF11, tcell.KeyF12:

			w.goToPanel(int(ev.Key() - tcell.KeyF1))
		case tcell.KeyEscape:
			if w.panelsOpen {
				w.SetPanelsOpen(false)
			} else {
				close(quit)
				return true
			}
		case tcell.KeyCtrlC:
			close(quit)
			return true
		case tcell.KeyCtrlL:
			w.resizeAndRedraw()
			return false
		}
	case *tcell.EventMouse:
		buttons := ev.Buttons()
		if buttons&tcell.ButtonPrimary != 0 {
			_, buttonY := ev.Position()
			for i, panel := range w.BottomPanels {
				_, panelY := panel.Position()
				if buttonY == panelY {
					w.goToPanel(i)
				}
			}
			// do not continue here so the now active panel can handle this event as well
		}
	case *tcell.EventResize:
		w.resizeAndRedraw()
		// TODO: maybe change this in the future and let it trickle down
		// instead of calling resize() manually
		return false
	case *EventPopupStateChanged:
		w.Render()
		return false
	case *EventPressedEnterInInputField:
		w.SetPanelsOpen(false)
		// don't continue here so that view can handle this as well
	default:
		// log.Printf("Event: %T, %+v", ev, ev)
	}

	if w.mainView.HandleEvent(ev) {
		return false
	}

	if w.statusbar.HandleEvent(ev) {
		return false
	}

	return false
}

func (w *Window) openPanelsOrPanelSelection() {
	// if panels are currently closed but at least one panel exists
	// already, then just open the existing panels, don't open
	// panel selection
	if w.activePanel != nil && !w.statusbar.panelsOpen {
		w.SetPanelsOpen(true)
		return
	} else {
		// if w.popup != nil { // TODO: weird
		// 	w.popup.SetActive(false)
		// }
		w.popup = w.panelSelection
		w.popup.SetActive(true)
		w.popup.SetVisible(true)
	}
}

func (w *Window) CreateAndAddPanel(panelType config.FilterType) {
	if w.activePanel != nil && !w.panelsOpen {
		return
	}

	w.AddPanel(NewPanel(panelType))
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

	panelsPlusStatusbarHeight := w.totalPanelHeight() + w.statusbar.Height()
	// TODO: in case the terminal gets to small, the panels will never
	// open again!
	if panelsPlusStatusbarHeight >= height {
		w.SetPanelsOpen(false)
		w.mainView.Resize(0, 0, width, height-w.statusbar.Height())
	} else {
		w.mainView.Resize(0, 0, width, height-panelsPlusStatusbarHeight)
		y := height - panelsPlusStatusbarHeight
		for _, p := range w.BottomPanels {
			p.Resize(0, y, width, 0) // x and height ignored
			y += p.Height()
		}
	}
	w.statusbar.Resize(0, height-1, width, 0) // x and height ignored
	w.exPanel.Resize(0, 3, width, 0)          // height ignored

	if w.popup != nil {
		w.popup.Resize(0, 0, 0, 0)
	}
}

func (w *Window) totalPanelHeight() int {
	if !w.panelsOpen {
		return 0
	}

	var totalPanelHeight = 0

	for _, p := range w.BottomPanels {
		totalPanelHeight += p.Height()
	}

	return totalPanelHeight
}

func (w *Window) AddPanel(newPanel components.Panel) error {
	// TODO: return error if total height of panels would exceed screen height
	w.BottomPanels = append(w.BottomPanels, newPanel)
	components.Add(newPanel, 1)
	w.SetActivePanel(newPanel)

	// resize() doesn't sound right here but will actually recalculate where
	// the panels should be placed and how big they are.
	// Is this really necessary?
	// w.resize()
	// w.Render()
	return nil
}

func (w *Window) RemovePanel() error {
	if len(w.BottomPanels) == 1 {
		return fmt.Errorf("cannot remove last panel")
	}

	var newActivePanel components.Panel
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

func (w *Window) SetActivePanel(p components.Panel) {
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

func (w *Window) goToPanel(no int) error {
	if no < 0 || no >= len(w.BottomPanels) {
		return fmt.Errorf("no panel at index %d", no)
	}

	w.SetActivePanel(w.BottomPanels[no])
	// It would probably be more natural to call render within the SetActivePanel()
	// (or even the individual SetActive() methods of the panels and InputFields),
	// but this way we avoid unnecessary redraws when switching panels.
	w.Render()

	return nil
}

func (w *Window) switchPanel(offset int) error {
	newPanelIndex := w.activePanelIndex() + offset

	if newPanelIndex < 0 || newPanelIndex >= len(w.BottomPanels) {
		return fmt.Errorf("no panel at index %d", newPanelIndex)
	}

	return w.goToPanel(newPanelIndex)
}

func (w *Window) SetPanelsOpen(panelsOpen bool) {
	w.panelsOpen = panelsOpen
	if panelsOpen {
		for _, p := range w.BottomPanels {
			p.SetVisible(true)
		}
		w.activePanel.SetActive(true)
	} else {
		for _, p := range w.BottomPanels {
			p.SetVisible(false)
			p.SetActive(false)
		}
	}
	GetScreen().PostEvent(NewEventPanelStateChanged(panelsOpen))
	w.resizeAndRedraw()
}

func (w *Window) PanelsOopen() bool {
	return w.panelsOpen
}
