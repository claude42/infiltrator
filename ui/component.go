package ui

type Component interface {
	SetActive(active bool)
	IsActive() bool
}

type ComponentImpl struct {
	active bool
}

func (c *ComponentImpl) SetActive(active bool) {
	c.active = active
}

func (c *ComponentImpl) IsActive() bool {
	return c.active
}
