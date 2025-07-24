package ui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

var FilterColors = []tcell.Color{
	tcell.ColorBrown,
	tcell.ColorMaroon,
	tcell.ColorGreen,
	tcell.ColorOlive,
	tcell.ColorNavy,
	tcell.ColorPurple,
	tcell.ColorTeal,
	tcell.ColorAliceBlue,
	tcell.ColorAquaMarine,
	tcell.ColorAzure,
	tcell.ColorBeige,
	tcell.ColorBisque,
	tcell.ColorBlack,
	tcell.ColorBlanchedAlmond,
	tcell.ColorBlue,
}

type colorManager struct {
	colors []colorMap
}

type colorMap struct {
	panel      Panel
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

func (c *colorManager) findNextUnassigendColorIndex() (uint8, error) {
	// 0 is used as "no specific color"
	for index := 1; index < len(FilterColors); index++ {
		if index >= 256 {
			return 0, fmt.Errorf("no unassigned color index found, maximum is 256")
		}
		found := false
		for _, cm := range c.colors {
			if cm.colorIndex == uint8(index) {
				found = true
				break
			}
		}
		if !found {
			return uint8(index), nil
		}
	}
	return 0, fmt.Errorf("no unassigned color index found")
}

func (c *colorManager) Add(panel Panel) uint8 {
	if panel == nil {
		log.Panic("Add() called with nil panel")
		return 0
	}

	unassigned, err := c.findNextUnassigendColorIndex()
	if err != nil {
		// TODO probably should not panic here, but handle the error gracefully
		log.Panic("Add() called with no unassigned color index available")
		return 0
	}

	c.colors = append(c.colors, colorMap{panel: panel, colorIndex: unassigned})

	return unassigned
}

func (c *colorManager) Remove(panel Panel) {
	for i, cm := range c.colors {
		if cm.panel == panel {
			c.colors = append(c.colors[:i], c.colors[i+1:]...)
			return
		}
	}
	log.Panicf("Remove() called with panel %v that is not registered", panel)
}

func (c *colorManager) GetColor(panel Panel) tcell.Color {
	for i, cm := range c.colors {
		if cm.panel == panel {
			if i < len(FilterColors) {
				return FilterColors[i]
			}
			return tcell.ColorDefault // fallback color
		}
	}
	return tcell.ColorDefault // fallback color if panel not found
}
