package ui

import (
	"log"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/fail"
	"github.com/gdamore/tcell/v2"
)

var FilterColors = [][2]tcell.Color{
	{tcell.ColorGreen, tcell.ColorGreen}, // just to detect wither something went wrong, should be tcell.ColorDefault
	{tcell.ColorRed, tcell.ColorDarkRed},
	{tcell.ColorLightPink, tcell.ColorPink},
	{tcell.ColorGoldenrod, tcell.ColorDarkGoldenrod},
	{tcell.ColorGreen, tcell.ColorDarkGreen},
	{tcell.ColorMaroon, tcell.ColorDarkMagenta},
	{tcell.ColorSalmon, tcell.ColorDarkSalmon},
	{tcell.ColorSlateBlue, tcell.ColorDarkSlateBlue},
	{tcell.ColorViolet, tcell.ColorDarkViolet},
	{tcell.ColorTurquoise, tcell.ColorDarkTurquoise},
	{tcell.ColorOrchid, tcell.ColorDarkOrchid},
	{tcell.ColorOlive, tcell.ColorDarkOliveGreen},
	{tcell.ColorKhaki, tcell.ColorDarkKhaki},
	{tcell.ColorOrange, tcell.ColorDarkOrange},
}

type colorManager struct {
	colors []colorMap
}

type colorMap struct {
	panel      components.Panel
	colorIndex uint8
}

var instance *colorManager

func GetColorManager() *colorManager {
	// not thread safe
	if instance == nil {
		instance = &colorManager{}
	}
	return instance
}

func (c *colorManager) findNextUnassigendColorIndex() uint8 {
	// 0 is used as "no specific color"
	for index := 1; index < len(FilterColors); index++ {
		fail.If(index >= 256, "No unassigned color index found, maximum is 256")
		found := false
		for _, cm := range c.colors {
			if cm.colorIndex == uint8(index) {
				found = true
				break
			}
		}
		if !found {
			return uint8(index)
		}
	}
	log.Panic("No unassigned color index found, maximum is 256")
	return 0
}

func (c *colorManager) Add(panel components.Panel) uint8 {
	fail.IfNil(panel, "Add() called with nil panel)")

	unassigned := c.findNextUnassigendColorIndex()

	c.colors = append(c.colors, colorMap{panel: panel, colorIndex: unassigned})

	return unassigned
}

func (c *colorManager) Remove(panel components.Panel) {
	for i, cm := range c.colors {
		if cm.panel == panel {
			c.colors = append(c.colors[:i], c.colors[i+1:]...)
			return
		}
	}
	// Don't fail in this case - will save us some housekeeping when destroying panels
	// log.Panicf("Remove() called with panel %v that is not registered", panel)
}

func (c *colorManager) Replace(oldPanel, newPanel components.Panel) {
	fail.IfNil(oldPanel, "ReplacePanel() called with nil oldPanel)")
	fail.IfNil(newPanel, "ReplacePanel() called with nil newPanel)")

	for i, cm := range c.colors {
		if cm.panel == oldPanel {
			c.colors[i].panel = newPanel
			return
		}
	}
	log.Panicf("Replace() called with old panel %v that is not registered", oldPanel)
}

func (c *colorManager) GetColor(panel components.Panel) [2]tcell.Color {
	for i, cm := range c.colors {
		if cm.panel == panel {
			if i < len(FilterColors) {
				return FilterColors[i]
			}
			return [2]tcell.Color{tcell.ColorDefault, tcell.ColorDefault} // fallback color
		}
	}
	return [2]tcell.Color{tcell.ColorDefault, tcell.ColorDefault} // fallback color if panel not found
}
