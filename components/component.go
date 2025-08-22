package components

import (
	"github.com/claude42/infiltrator/util"
)

type Component interface {
	SetActive(active bool)
	IsActive() bool
	SetVisible(visible bool)
	IsVisible() bool
	Show()
	Hide()
	Resize(x, y, width, height int)
	Render(updateScreen bool)
	Position() (int, int)
	Size() (int, int)
	Width() int
	Height() int

	util.EventHandler
	util.Observable
}

type ComponentImpl struct {
	util.EventHandlerIgnoreImpl
	util.ObservableIgnore
	x, y, width, height int

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

func (c *ComponentImpl) Show() {
	c.active = true
	c.visible = true
}

func (c *ComponentImpl) Hide() {
	c.active = false
	c.visible = false
}

func (c *ComponentImpl) Resize(x, y, width, height int) {
	if x != -1 {
		c.x = x
	}
	if y != -1 {
		c.y = y
	}
	if width != -1 {
		c.width = width
	}

	if height != -1 {
		c.height = height
	}
}

func (c *ComponentImpl) Position() (int, int) {
	return c.x, c.y
}

func (c *ComponentImpl) Size() (int, int) {
	return c.width, c.height
}

func (c *ComponentImpl) Width() int {
	return c.width
}

func (c *ComponentImpl) Height() int {
	return c.height
}
