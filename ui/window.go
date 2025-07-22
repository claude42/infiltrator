package ui

import (
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
	activePanel  int
}

func Setup(pipeline *model.Pipeline) *Window {
	if window != nil {
		log.Fatalln("ui.setup() called twice!")
	}

	window = &Window{}

	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	window.SetView(NewView(pipeline))
	k := NewKeywordPanel()
	window.AddPanel(k)

	/*t := NewTextEntryPanel()
	t.SetContent("hurz")
	window.AddPanel(t)*/
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

		if w.mainView.HandleEvent(ev) {
			continue
		}

		if w.BottomPanels[w.activePanel].HandleEvent(ev) {
			continue
		}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q':
					close(quit)
					return
				}
			case tcell.KeyBacktab:
				w.switchPanel(-1)
			case tcell.KeyTab:
				w.switchPanel(1)
			case tcell.KeyEscape, tcell.KeyCtrlC:
				close(quit)
				return
			case tcell.KeyCtrlL:
				w.resizeAndRedraw()
			}
		case *tcell.EventResize:
			w.resizeAndRedraw()
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
			y += p.GetHeight()
		}
	}
}

func (w *Window) totalPanelHeight() int {
	if !w.PanelsOpen {
		return 0
	}

	var totalPanelHeight = 0

	for _, p := range w.BottomPanels {
		totalPanelHeight += p.GetHeight()
	}

	return totalPanelHeight
}

func (w *Window) AddPanel(newPanel Panel) {
	w.BottomPanels = append(w.BottomPanels, newPanel)
	w.SetActivePanel(len(w.BottomPanels) - 1)
}

func (w *Window) SetActivePanel(pos int) {
	// should really happen but just in case
	if w.activePanel < 0 || w.activePanel >= len(w.BottomPanels) {
		log.Fatalf("Spurios curent panel number %d", w.activePanel)
	}

	w.BottomPanels[w.activePanel].SetActive(false)

	w.activePanel = pos
	w.BottomPanels[w.activePanel].SetActive(true)

	// is there any case where the whole window (instead of the affected panel)
	// would have to be redrawn?
	// w.Render()
}

func (w *Window) switchPanel(offset int) {
	newPanelIndex := w.activePanel + offset

	if newPanelIndex < 0 || newPanelIndex >= len(w.BottomPanels) {
		screen.Beep()
		return
	}

	w.SetActivePanel(newPanelIndex)
}

func (w *Window) SetView(v *View) {
	window.mainView = v
}

// As soon as we get more options, we should use a struct for this
func (w *Window) ShowLineNumbers(showLineNumbers bool) {
	w.mainView.SetShowLineNumbers(showLineNumbers)
}
