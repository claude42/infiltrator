package components

import (
	"sync"

	"github.com/claude42/infiltrator/fail"
	"github.com/gdamore/tcell/v2"
)

type Container interface {
	Component

	Add(component Component)
	Remove(component Component)
}

type ContainerImpl struct {
	ComponentImpl

	once      sync.Once
	contained map[Component]struct{}
}

func (c *ContainerImpl) Add(component Component) {
	c.once.Do(func() {
		c.contained = make(map[Component]struct{})
	})
	c.contained[component] = struct{}{}
	component.Watch(c)
}

func (c *ContainerImpl) Remove(component Component) {
	fail.IfNil(c.contained, "ContainerImpl: Remove called on uninitialized container")
	delete(c.contained, component)
	component.Unwatch(c)
}

func (c *ContainerImpl) Contained() map[Component]struct{} {
	return c.contained
}

func (c *ContainerImpl) SetActive(active bool) {
	c.ComponentImpl.SetActive(active)
	for c := range c.contained {
		c.SetActive(active)
	}
}

func (c *ContainerImpl) SetVisible(visible bool) {
	c.ComponentImpl.SetVisible(visible)
	for c := range c.contained {
		c.SetVisible(visible)
	}
}

func (c *ContainerImpl) Show() {
	c.SetActive(true)
	c.SetVisible(true)
}

func (c *ContainerImpl) Hide() {
	c.SetActive(false)
	c.SetVisible(false)
}

func (c *ContainerImpl) Render(updateScreen bool) {
	for i := range c.contained {
		i.Render(false)
	}

	if updateScreen {
		Screen.Show()
	}
}

func (c *ContainerImpl) HandleEvent(ev tcell.Event) bool {
	if !c.IsActive() {
		return false
	}

	for i := range c.contained {
		if i.HandleEvent(ev) {
			return true
		}
	}

	return false
}
