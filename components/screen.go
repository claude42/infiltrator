package components

import (
	"cmp"
	"slices"

	"github.com/gdamore/tcell/v2"
)

var Screen tcell.Screen

func Beep() {
	Screen.Beep()
}

type componentAndZ struct {
	component Component
	z         int
}

var stackedComponents []componentAndZ

func Add(component Component, z int) {
	stackedComponents = append(stackedComponents,
		componentAndZ{component, z})

	slices.SortFunc(stackedComponents, func(a, b componentAndZ) int {
		return cmp.Compare(a.z, b.z)
	})
	RenderAll(false)
	Screen.Show()
}

func Remove(component Component) {
	for i, c := range stackedComponents {
		if c.component == component {
			stackedComponents = append(stackedComponents[:i], stackedComponents[i+1:]...)
			return
		}
	}
	RenderAll(false)
	Screen.Show()
}

func RenderAll(updateScreen bool) {
	for _, c := range stackedComponents {
		c.component.Render(false)
	}

	if updateScreen {
		Screen.Show()
	}
}

func HandleEventAll(ev tcell.Event) bool {
	// make a copy b/c the original slice gets potentially modified
	stackedComponentsCopy := make([]componentAndZ, len(stackedComponents))
	copy(stackedComponentsCopy, stackedComponents)

	for i := len(stackedComponentsCopy) - 1; i >= 0; i-- {
		if stackedComponentsCopy[i].component.HandleEvent(ev) {
			// log.Printf("Event\n%+v\nconsumed by\n%+v", ev, stackedComponentsCopy[i].component)
			return true
		}
	}
	return false
}
