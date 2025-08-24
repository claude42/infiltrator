package ui

import (
	"context"
	"fmt"
	"log"
	"os"
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
	window.mainView.Show()

	window.statusbar = NewStatusbar()
	components.Add(window.statusbar, 0)
	window.statusbar.SetVisible(true)

	window.panelSelection = NewPanelSelection()
	components.Add(window.panelSelection, 2)

	window.exPanel = NewExPanel()
	components.Add(window.exPanel, 1)
	window.exPanel.SetContent("Hallo")
	window.exPanel.SetPrompt("Testerli")
	// window.exPanel.SetVisible(true)
	// window.exPanel.SetActive(true)

	setupScreen()

	// ShowQuestionBar("Hello: ", "Claude", func(value string) {
	// 	log.Printf("Did it %s", value)
	// })

	// ShowYesNoBar("todalo", "Hell (yes/no)?", func(name string) {
	// 	log.Printf("You did it %s", name)
	// }, func(name string) {
	// 	log.Printf("Maybe next time %s", name)
	// })

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
	components.RenderAll(true)
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
	// log.Printf("Main Loop: %T", ev)

	if kev, ok := ev.(*tcell.EventKey); ok {
		log.Printf("Key Name: %s", kev.Name())
	}

	if components.HandleEventAll(ev) {
		return false
	}

	switch ev := ev.(type) {
	case *tcell.EventKey:
		// log.Print("-------------------------------------------------------")
		// log.Printf("Modmast: %d", ev.Modifiers())
		// log.Printf("Rune: %c", ev.Rune())
		// log.Printf("Key: %s", tcell.KeyNames[ev.Key()])

		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case '/':
				GetPanelManager().openPanelsOrPanelSelection()
				w.Render()
				return false
			case 'q':
				quit <- "Good bye!"
				close(quit)
				return true
			case 'S':
				w.savePreset()
				return false
			}
		case tcell.KeyBacktab:
			err := GetPanelManager().switchPanel(-1)
			if err != nil {
				screen.Beep()
			}
			return false
		case tcell.KeyTab:
			err := GetPanelManager().switchPanel(1)
			if err != nil {
				screen.Beep()
			}
			return false
		case tcell.KeyCtrlP:
			GetPanelManager().openPanelsOrPanelSelection()
			w.Render()
			return false
		case tcell.KeyCtrlO:
			toBeDestroyed := GetPanelManager().activePanel
			err := GetPanelManager().Remove()
			if err != nil {
				screen.Beep()
			}
			components.Remove(toBeDestroyed)
			w.Render()
			return false
		case tcell.KeyF1, tcell.KeyF2, tcell.KeyF3, tcell.KeyF4, tcell.KeyF5, tcell.KeyF6,
			tcell.KeyF7, tcell.KeyF8, tcell.KeyF9, tcell.KeyF10, tcell.KeyF11, tcell.KeyF12:

			GetPanelManager().goTo(int(ev.Key() - tcell.KeyF1))
		case tcell.KeyEscape:
			if GetPanelManager().panelsOpen {
				GetPanelManager().SetPanelsOpen(false)
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
			for i, panel := range GetPanelManager().panels {
				_, panelY := panel.Position()
				if buttonY == panelY {
					GetPanelManager().goTo(i)
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
		GetPanelManager().SetPanelsOpen(false)
		// don't continue here so that view can handle this as well
	default:
		// log.Printf("Event: %T, %+v", ev, ev)
	}

	return false
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

	panelsPlusStatusbarHeight := GetPanelManager().totalHeight() + w.statusbar.Height()
	// TODO: in case the terminal gets to small, the panels will never
	// open again!
	if panelsPlusStatusbarHeight >= height {
		GetPanelManager().SetPanelsOpen(false)
		w.mainView.Resize(0, 0, width, height-w.statusbar.Height())
	} else {
		w.mainView.Resize(0, 0, width, height-panelsPlusStatusbarHeight)
		y := height - panelsPlusStatusbarHeight
		for _, p := range GetPanelManager().panels {
			p.Resize(0, y, width, 0) // x and height ignored
			y += p.Height()
		}
	}
	w.statusbar.Resize(0, height-1, width, 0) // x and height ignored
	w.exPanel.Resize(0, height-1, width, 0)   // height ignored

	if w.popup != nil {
		w.popup.Resize(-1, -1, -1, -1)
	}
}

func (w *Window) CreatePresetPanels() {
	for _, panelConfig := range config.Panels() {
		GetPanelManager().Add(NewPanelWithConfig(&panelConfig))
	}
	w.resizeAndRedraw()
}

func (w *Window) savePreset() {
	ShowQuestionBar("Preset name: ", config.User().Preset, func(presetName string) {
		presetFileName := config.BuildFullPresetPath(presetName)

		_, err := os.Stat(presetFileName)
		if err != nil {
			config.WritePreset(presetFileName)
		}

		ShowYesNoBar("File exists! Overwrite (y/n)?", func() {
			GetPanelManager().copyPanelsToConfig()
			config.WritePreset(presetFileName)
		}, nil)
	})
}
