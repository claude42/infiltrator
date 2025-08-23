package components

import (
	"slices"
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
	contained []Component
}

func (c *ContainerImpl) Add(component Component) {
	c.once.Do(func() {
		c.contained = make([]Component, 0)
	})
	c.contained = append(c.contained, component)
	component.Watch(c)
}

func (c *ContainerImpl) Remove(component Component) {
	fail.IfNil(c.contained, "ContainerImpl: Remove called on uninitialized container")

	i := slices.Index(c.contained, component)
	if i < 0 {
		return
	}
	c.contained = append(c.contained[:i], c.contained[i+1:]...)
	component.Unwatch(c)
}

func (c *ContainerImpl) Contained() []Component {
	return c.contained
}

func (c *ContainerImpl) SetActive(active bool) {
	c.ComponentImpl.SetActive(active)
	for _, i := range c.contained {
		i.SetActive(active)
	}
}

func (c *ContainerImpl) SetVisible(visible bool) {
	c.ComponentImpl.SetVisible(visible)
	for _, i := range c.contained {
		i.SetVisible(visible)
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
	for _, i := range c.contained {
		i.Render(false)
	}

	if updateScreen {
		Screen.Show()
	}
}

func (c *ContainerImpl) HandleEvent(ev tcell.Event) bool {
	for _, i := range c.contained {
		if i.HandleEvent(ev) {
			return true
		}
	}

	return false
}
