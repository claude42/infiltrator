package ui

import (
	"fmt"
	"log"

	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type View struct {
	pipeline              *model.Pipeline
	viewWidth, viewHeight int
	curX, curY            int
	showLineNumbers       bool
	followFile            bool

	ComponentImpl
}

func NewView(pipeline *model.Pipeline) *View {
	v := &View{}
	v.viewWidth, v.viewHeight = screen.Size()

	v.SetPipeline(pipeline)

	return v
}

func (v *View) SetPipeline(pipeline *model.Pipeline) {
	v.pipeline = pipeline
	pipeline.Watch(v)
	v.SetCursor(0, 0)
}

func (v *View) Render(updateScreen bool) {
	screenBuffer := v.pipeline.ScreenBuffer(v.curY, v.viewHeight)
	// if the first line on the screen doesn't match, adjust curY, so
	// e.g. the next cursor down will have an effect
	v.curY = util.IntMax(screenBuffer[0].No, 0)

	for y := 0; y < v.viewHeight; y++ {
		v.renderLine(screenBuffer[y], y)
	}

	if updateScreen {
		screen.Show()
	}
}

func (v *View) renderLine(line model.Line, y int) {
	str := line.Str
	var start = 0

	if v.showLineNumbers {
		start = v.renderLineNumber(line, y)
	}

	for x := start; x < v.viewWidth; x++ {
		var r rune = ' '
		var lineXPos = v.curX + x - start
		var style tcell.Style

		switch line.Status {
		case model.LineWithoutStatus, model.LineMatched:
			style = ViewStyle
		case model.LineDimmed:
			style = ViewDimmedStyle
		}

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

func (v *View) renderLineNumber(line model.Line, y int) int {
	if line.No < 0 {
		return 0 // TODO: 0 ok?
	}
	_, length, err := v.pipeline.Size()
	if err != nil {
		return 0 // TODO: 0 ok?
	}

	str := fmt.Sprintf("%*d ", util.CountDigits(length-1), line.No)

	var x int
	var style tcell.Style
	switch line.Status {
	case model.LineWithoutStatus, model.LineMatched:
		style = ViewLineNumberStyle
	case model.LineDimmed:
		style = ViewDimmedLineNumberStyle
		// case model.LineHidden:
		// todo
	}
	for x = 0; x < v.viewWidth && x < len(str); x++ {
		screen.SetContent(x, y, rune(str[x]), nil, style)
	}

	return x
}

func (v *View) SetCursor(x, y int) error {
	var err error

	if v.curY != y {
		v.pipeline.InvalidateScreenBuffer()
	}

	v.curX, v.curY, err = v.stayWithinLimits(x, y)

	return err
}

func (v *View) GetCursor() (x, y int) {
	return v.curX, v.curY
}

func (v *View) Move(xOff, yOff int) {
	var err error
	v.curX, v.curY, err = v.stayWithinLimits(v.curX+xOff, v.curY+yOff)

	if err != nil {
		screen.Beep()
	}
	v.Render(true)
}

func (v *View) stayWithinLimits(x int, y int) (int, int, error) {
	var newX, newY int
	var errX, errY error
	var err error = nil

	width, length, err := v.pipeline.Size()
	if err != nil {
		return 0, 0, err
	}

	newX, errX = util.InBetween(x, 0, width)
	newY, errY = util.InBetween(y, 0, length-1)

	if errX != nil {
		err = errX
	} else if errY != nil {
		err = errY
	}

	return newX, newY, err
}

func (v *View) Resize(x, y, width, height int) {
	// x, y ignored for now
	v.viewWidth = width
	v.viewHeight = height
	v.pipeline.RefreshScreenBuffer(v.curY, v.viewHeight)
}

func (v *View) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *model.EventBufferDirty:
		v.reactToFileUpdate()
		return true
	case *tcell.EventKey:
		log.Printf("View.HandleEvent: %s", ev.Name())
		switch ev.Key() {
		case tcell.KeyDown:
			err := v.scrollDown(true)
			if err != nil {
				screen.Beep()
			}
			return true
		case tcell.KeyUp:
			err := v.scrollUp(true)
			if err != nil {
				screen.Beep()
			}
			return true

		case tcell.KeyRight:
			err := v.scrollHorizontal(1)
			if err != nil {
				screen.Beep()
			}
			return true
		case tcell.KeyLeft:
			err := v.scrollHorizontal(-1)
			if err != nil {
				screen.Beep()
			}
			return true
		case tcell.KeyCtrlF, tcell.KeyPgDn:
			err := v.pageDown()
			if err != nil {
				screen.Beep()
			}
			return true
		case tcell.KeyCtrlB, tcell.KeyPgUp:
			err := v.pageUp()
			if err != nil {
				screen.Beep()
			}
			return true
		case tcell.KeyCtrlA, tcell.KeyHome:
			err := v.scrollHome()
			if err != nil {
				screen.Beep()
			}
			return true
		case tcell.KeyCtrlE, tcell.KeyEnd:
			err := v.scrollEnd()
			if err != nil {
				screen.Beep()
			}
			return true
		}
	case *tcell.EventMouse:
		buttons := ev.Buttons()
		log.Printf("Wheel: %d", buttons)

		// Horizontal mouse wheel doesn't seem to work with the terminals I
		// have access to but we'll leave it in anyways...
		if buttons&tcell.WheelUp != 0 {
			v.scrollUp(true)
			return true
		} else if buttons&tcell.WheelDown != 0 {
			v.scrollDown(true)
			return true
		} else if buttons&tcell.WheelLeft != 0 {
			v.scrollHorizontal(-1)
			return true
		} else if buttons&tcell.WheelRight != 0 {
			v.scrollHorizontal(1)
			return true
		}
	case *model.EventFilterOutput:
		v.Render(true)
	}

	return false
}

func (v *View) reactToFileUpdate() {
	v.pipeline.InvalidateScreenBuffer()
	if v.followFile {
		v.scrollEnd()
	} else {
		v.Render(true)
	}
}

func (v *View) scrollUp(render bool) error {
	err := v.pipeline.ScrollUpLineBuffer()
	if err != nil {
		log.Printf("Eventloop ScrollUpLineBufferError")
		return err
	}
	if render {
		v.Render(true)
	}
	return nil
}

func (v *View) pageUp() error {
	var err error
	for i := 0; i < v.viewHeight-1; i++ {
		err = v.scrollUp(false)
		if err != nil {
			break
		}
	}
	v.Render(true)
	return err
}

func (v *View) scrollDown(render bool) error {
	err := v.pipeline.ScrollDownLineBuffer()
	if err != nil {
		log.Printf("Eventloop ScrollUpLineBufferError")
		return err
	}
	if render {
		v.Render(true)
	}
	return nil
}

func (v *View) pageDown() error {
	var err error
	for i := 0; i < v.viewHeight-1; i++ {
		err = v.scrollDown(false)
		if err != nil {
			break
		}
	}
	v.Render(true)
	return err
}

func (v *View) scrollHorizontal(offset int) error {
	width, _, err := v.pipeline.Size()
	if err != nil {
		return err
	}

	newX, err := util.InBetween(v.curX+offset, 0, width)
	if err != nil {
		return err
	}

	v.curX = newX
	v.Render(true)
	return nil
}

func (v *View) scrollHome() error {
	v.SetCursor(v.curX, 0)
	v.Render(true)

	return nil
}

func (v *View) scrollEnd() error {
	_, length, err := v.pipeline.Size()
	if err != nil {
		return err
	}

	v.SetCursor(v.curX, length-v.viewHeight)
	v.Render(true)

	return nil
}

func (v *View) SetShowLineNumbers(showLineNumbers bool) {
	v.showLineNumbers = showLineNumbers
}

func (v *View) SetFollowFile(followFile bool) {
	v.followFile = followFile
	v.scrollEnd()
}
