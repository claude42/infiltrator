package ui

import (
	"github.com/claude42/infiltrator/util"
)

type Component interface {
	SetActive(active bool)
	IsActive() bool
	Resize(x, y, width, height int)
	Render(updateScreen bool)

	util.EventHandler
}

type ComponentImpl struct {
	util.EventHandlerPanicImpl

	active bool
}

func (c *ComponentImpl) SetActive(active bool) {
	c.active = active
}

func (c *ComponentImpl) IsActive() bool {
	return c.active
}
