package components

import (
	"sync"
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
}

func (c *ContainerImpl) Remove(component Component) {
	delete(c.contained, component)
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
