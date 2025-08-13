package components

import (
	"github.com/claude42/infiltrator/util"
)

type Component interface {
	SetActive(active bool)
	IsActive() bool
	SetVisible(visible bool)
	IsVisible() bool
	Resize(x, y, width, height int)
	Render(updateScreen bool)

	util.EventHandler
}

type ComponentImpl struct {
	util.EventHandlerPanicImpl

	active  bool
	visible bool
}

func (c *ComponentImpl) SetActive(active bool) {
	c.active = active
}

func (c *ComponentImpl) IsActive() bool {
	return c.active
}

func (c *ComponentImpl) SetVisible(visible bool) {
	c.visible = visible
}

func (c *ComponentImpl) IsVisible() bool {
	return c.visible
}
