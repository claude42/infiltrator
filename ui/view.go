package ui

import (
	"fmt"
	"log"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type View struct {
	ComponentImpl

	viewWidth, viewHeight int
	curX, curY            int
	CurrentDisplay        *model.Display
}

func NewView() *View {
	v := &View{}

	return v
}

func (v *View) Render(display *model.Display, updateScreen bool) {
	log.Printf("view.Render()")
	// display will be nil if called from Windows.Render(). Only go ahead
	// if there's already a v.currentDisplay
	if display == nil && v.CurrentDisplay == nil {
		return
	}

	if display != nil {
		v.CurrentDisplay = display
	}

	// no sense in rendering nothing
	if len(v.CurrentDisplay.Buffer) == 0 {
		return
	}

	// skip rendering if display buffer and viewheight differ. Most likely
	// backend Goroutine needs to catch up
	if len(v.CurrentDisplay.Buffer) != v.viewHeight {
		return
	}

	// If what the View thinks is the first line displayed on screen vs.
	// what's in the buffer, adjust curY
	v.curY = util.IntMax(v.CurrentDisplay.Buffer[0].No, 0)

	// Guard clause above should catch this, but still: Make sure we don't
	// render beyond v.viewHeight!
	for y := 0; y < len(v.CurrentDisplay.Buffer) && y < v.viewHeight; y++ {
		v.renderLine(v.CurrentDisplay.Buffer[y], y)
	}

	if updateScreen {
		screen.Show()
	}
}

func (v *View) renderLine(line model.Line, y int) {
	str := line.Str
	var start = 0
	matched := line.No == v.CurrentDisplay.CurrentMatch

	if config.GetConfiguration().ShowLineNumbers {
		start = v.renderLineNumber(line, y, matched)
	}

	lineStyle := v.determineStyle(line, matched)

	for x := start; x < v.viewWidth; x++ {
		var r rune = ' '
		style := lineStyle
		var lineXPos = v.curX + x - start

		if v.curX+x < len(str)+start {
			r = rune(str[lineXPos])

			// in case we're on the last screen column, render an inverse '>'
			if x == v.viewWidth-1 && v.curX+x+1 < len(str) {
				r = '>'
				style = style.Reverse(true)
			} else if line.ColorIndex[lineXPos] > 0 {
				switch line.Status {
				case model.LineWithoutStatus, model.LineMatched:
					style = style.Foreground(FilterColors[line.ColorIndex[lineXPos]][0])
				case model.LineDimmed:
					style = style.Foreground(FilterColors[line.ColorIndex[lineXPos]][1])
				}
				style = style.Reverse(true)
			}
		}

		screen.SetContent(x, y, r, nil, style)
	}
}

func (v *View) determineStyle(line model.Line, matched bool) tcell.Style {
	if matched {
		return CurrentMatchStyle
	} else {
		switch line.Status {
		case model.LineWithoutStatus, model.LineMatched:
			return ViewStyle
		case model.LineDimmed:
			return ViewDimmedStyle
		default:
			// should only occur for hidden lines and therefore not matter
			// but let's use a distinctive color to spot any errors
			// immediately
			return DefStyle.Foreground(tcell.ColorGreen)
		}
	}
}

func (v *View) renderLineNumber(line model.Line, y int, matched bool) int {
	if line.No < 0 {
		return 0 // TODO: 0 ok?
	}

	str := fmt.Sprintf("%*d ", util.CountDigits(v.CurrentDisplay.TotalLength-1), line.No)

	var x int
	style := v.determineLineNumberStyle(line, matched)
	for x = 0; x < v.viewWidth && x < len(str); x++ {
		screen.SetContent(x, y, rune(str[x]), nil, style)
	}

	return x
}

func (v *View) determineLineNumberStyle(line model.Line, matched bool) tcell.Style {
	if matched {
		return ViewCurrentMatchLineNumberStyle
	} else {
		switch line.Status {
		case model.LineWithoutStatus, model.LineMatched:
			return ViewLineNumberStyle
		case model.LineDimmed:
			return ViewDimmedLineNumberStyle
		default:
			// should only occur for hidden lines and therefore not matter
			// but let's use a distinctive color to spot any errors
			// immediately
			return DefStyle.Foreground(tcell.ColorGreen)
		}
	}
}

func (v *View) GetCursor() (x, y int) {
	return v.curX, v.curY
}

func (v *View) Resize(x, y, width, height int) {
	// x, y ignored for now
	v.viewWidth = width
	v.viewHeight = height
	model.GetFilterManager().SetDisplayHeight(v.viewHeight)
	//model.GetFilterManager().RefreshScreenBuffer(v.curY, v.viewHeight)
}

func (v *View) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *model.EventDisplay:
		log.Printf("DisplayEvent. totalLength: %d, percentage: %d", ev.Display.TotalLength, ev.Display.Percentage)
		v.Render(&ev.Display, true)
	// completely handled in FileManager - no need to do something elaborate here
	// case *model.EventFileChanged:
	// 	log.Printf("EventFileChanged %d", ev.Length())
	// 	return false
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case '<', 'g':
				model.GetFilterManager().ScrollHome()
				return true
			case '>', 'G':
				model.GetFilterManager().ScrollEnd()
				return true
			case 'j':
				model.GetFilterManager().ScrollDown()
				return true
			case 'k':
				model.GetFilterManager().ScrollUp()
				return true
			case 'h':
				v.scrollHorizontal(-1)
				return true
			case 'l':
				v.scrollHorizontal(1)
				return true
			case ' ', 'f':
				model.GetFilterManager().PageDown()
				return true
			case 'b':
				model.GetFilterManager().PageUp()
			case 'n':
				model.GetFilterManager().FindMatch(1)
				return true
			case 'N':
				model.GetFilterManager().FindMatch(-1)
			case 'F':
				model.GetFilterManager().ToggleFollowMode()
			}
		case tcell.KeyDown, tcell.KeyEnter:
			model.GetFilterManager().ScrollDown()
			return true
		case tcell.KeyUp:
			model.GetFilterManager().ScrollUp()
			return true
		case tcell.KeyRight:
			v.scrollHorizontal(1)
			return true
		case tcell.KeyLeft:
			v.scrollHorizontal(-1)
			return true
		case tcell.KeyCtrlF, tcell.KeyPgDn:
			model.GetFilterManager().PageDown()
			return true
		case tcell.KeyCtrlB, tcell.KeyPgUp:
			model.GetFilterManager().PageUp()
			return true
		case tcell.KeyCtrlA, tcell.KeyHome:
			model.GetFilterManager().ScrollHome()
			return true
		case tcell.KeyCtrlE, tcell.KeyEnd:
			model.GetFilterManager().ScrollEnd()
			return true
		}
	case *tcell.EventMouse:
		buttons := ev.Buttons()
		log.Printf("Wheel: %d", buttons)

		// Horizontal mouse wheel doesn't seem to work with the terminals I
		// have access to but we'll leave it in anyways...
		if buttons&tcell.WheelUp != 0 {
			model.GetFilterManager().ScrollUp()
			return true
		} else if buttons&tcell.WheelDown != 0 {
			model.GetFilterManager().ScrollDown()
			return true
		} else if buttons&tcell.WheelLeft != 0 {
			v.scrollHorizontal(-1)
			return true
		} else if buttons&tcell.WheelRight != 0 {
			v.scrollHorizontal(1)
			return true
		}
	}

	return false
}

func (v *View) scrollHorizontal(offset int) {
	// width, _, err := model.GetFilterManager().Size()
	// if err != nil {
	// 	// TODO: rather fail than beep?
	// 	screen.Beep()
	// 	return
	// }

	// newX, err := util.InBetween(v.curX+offset, 0, width)
	// if err != nil {
	// 	screen.Beep()
	// 	return
	// }

	// v.curX = newX
	// v.Render(true)
}
