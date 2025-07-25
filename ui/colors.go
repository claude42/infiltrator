package ui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

var FilterColors = [][2]tcell.Color{
	{tcell.ColorGreen, tcell.ColorGreen}, // just to detect wither something went wrong
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

// var FilterColors = []tcell.Color{
// 	tcell.ColorBrown,
// 	tcell.ColorMaroon,
// 	tcell.ColorGreen,
// 	tcell.ColorOlive,
// 	tcell.ColorNavy,
// 	tcell.ColorPurple,
// 	tcell.ColorTeal,
// 	tcell.ColorAliceBlue,
// 	tcell.ColorAquaMarine,
// 	tcell.ColorAzure,
// 	tcell.ColorBeige,
// 	tcell.ColorBisque,
// 	tcell.ColorBlack,
// 	tcell.ColorBlanchedAlmond,
// 	tcell.ColorBlue,
// }

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

func (c *colorManager) Add(panel Panel) (uint8, error) {
	if panel == nil {
		log.Panic("Add() called with nil panel")
		return 0, nil
	}

	unassigned, err := c.findNextUnassigendColorIndex()
	if err != nil {

		return 0, fmt.Errorf("maximum number of panels")
	}

	c.colors = append(c.colors, colorMap{panel: panel, colorIndex: unassigned})

	return unassigned, nil
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

func (c *colorManager) GetColor(panel Panel) [2]tcell.Color {
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
