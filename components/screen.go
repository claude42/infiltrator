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
}

func Remove(component Component) {
	for i, c := range stackedComponents {
		if c.component == component {
			stackedComponents = append(stackedComponents[:i], stackedComponents[i+1:]...)
			return
		}
	}
}

func RenderAll(updateScreen bool) {
	for _, c := range stackedComponents {
		c.component.Render(updateScreen)
	}
}
