package ui

import (
	"fmt"

	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/reader"
	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type View struct {
	components.ComponentImpl

	viewWidth, viewHeight int
	CurrentDisplay        *model.Display
}

func NewView() *View {
	v := &View{}
	v.SetVisible(true)

	return v
}

func (v *View) Render(updateScreen bool) {
	v.RenderNewDisplay(nil, updateScreen)
}

func (v *View) RenderNewDisplay(display *model.Display, updateScreen bool) {
	if !v.IsVisible() {
		return
	}

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

	// Guard clause above should catch this, but still: Make sure we don't
	// render beyond v.viewHeight!
	for y := 0; y < len(v.CurrentDisplay.Buffer) && y < v.viewHeight; y++ {
		v.renderLine(v.CurrentDisplay.Buffer[y], y)
	}

	if updateScreen {
		screen.Show()
	}
}

func (v *View) renderLine(line *reader.Line, y int) {
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
		var lineXPos = v.CurrentDisplay.CurrentCol + x - start

		if v.CurrentDisplay.CurrentCol+x < len(str)+start {
			r = rune(str[lineXPos])

			// in case we're on the last screen column, render an inverse '>'
			if x == v.viewWidth-1 && v.CurrentDisplay.CurrentCol+x+1 < len(str) {
				r = '>'
				style = style.Reverse(true)
			} else if line.ColorIndex[lineXPos] > 0 {
				switch line.Status {
				case reader.LineWithoutStatus, reader.LineMatched:
					style = style.Foreground(FilterColors[line.ColorIndex[lineXPos]][0])
				case reader.LineDimmed:
					style = style.Foreground(FilterColors[line.ColorIndex[lineXPos]][1])
				}
				style = style.Reverse(true)
			}
		}

		screen.SetContent(x, y, r, nil, style)
	}
}

func (v *View) determineStyle(line *reader.Line, matched bool) tcell.Style {
	if matched {
		return CurrentMatchStyle
	} else {
		switch line.Status {
		case reader.LineWithoutStatus, reader.LineMatched:
			return ViewStyle
		case reader.LineDimmed:
			return ViewDimmedStyle
		default:
			// should only occur for hidden lines and therefore not matter
			// but let's use a distinctive color to spot any errors
			// immediately
			return DefStyle.Foreground(tcell.ColorGreen)
		}
	}
}

func (v *View) renderLineNumber(line *reader.Line, y int, matched bool) int {
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

func (v *View) determineLineNumberStyle(line *reader.Line, matched bool) tcell.Style {
	if matched {
		return ViewCurrentMatchLineNumberStyle
	} else {
		switch line.Status {
		case reader.LineWithoutStatus, reader.LineMatched:
			return ViewLineNumberStyle
		case reader.LineDimmed:
			return ViewDimmedLineNumberStyle
		default:
			// should only occur for hidden lines and therefore not matter
			// but let's use a distinctive color to spot any errors
			// immediately
			return DefStyle.Foreground(tcell.ColorGreen)
		}
	}
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
		// log.Printf("DisplayEvent. totalLength: %d, percentage: %d", ev.Display.TotalLength, ev.Display.Percentage)
		v.RenderNewDisplay(&ev.Display, true)
		return false
	case *model.EventError:
		if ev.Beep {
			screen.Beep()
			return true
		}
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
				model.GetFilterManager().ScrollHorizontal(-1)
				return true
			case 'l':
				model.GetFilterManager().ScrollHorizontal(1)
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
			model.GetFilterManager().ScrollHorizontal(1)
			v.CurrentDisplay.CurrentCol++
			v.RenderNewDisplay(nil, true)
			return true
		case tcell.KeyLeft:
			model.GetFilterManager().ScrollHorizontal(-1)
			if v.CurrentDisplay.CurrentCol > 0 {
				v.CurrentDisplay.CurrentCol--
			}
			v.RenderNewDisplay(nil, true)
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
		// log.Printf("Wheel: %d", buttons)

		// Horizontal mouse wheel doesn't seem to work with the terminals I
		// have access to but we'll leave it in anyways...
		if buttons&tcell.WheelUp != 0 {
			model.GetFilterManager().ScrollUp()
			return true
		} else if buttons&tcell.WheelDown != 0 {
			model.GetFilterManager().ScrollDown()
			return true
		} else if buttons&tcell.WheelLeft != 0 {
			model.GetFilterManager().ScrollHorizontal(-1)
			return true
		} else if buttons&tcell.WheelRight != 0 {
			model.GetFilterManager().ScrollHorizontal(1)
			return true
		}
	}

	return false
}
