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

	showLineNumbers bool

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
	log.Printf("Added view as eventhandler to pipeline")
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
	v.curX, v.curY, err = v.stayWithinLimits(x, y)
	return err
}

func (v *View) GetCursor() (x, y int) {
	return v.curX, v.curY
}

func (v *View) MoveCursor(xOff, yOff int) error {
	var err error
	v.curX, v.curY, err = v.stayWithinLimits(v.curX+xOff, v.curY+yOff)
	return err
}

func (v *View) Move(xOff, yOff int) {
	if v.MoveCursor(xOff, yOff) != nil {
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
	log.Printf("View.Handling: %T", ev)
	switch ev := ev.(type) {
	case *tcell.EventKey:
		log.Printf("Handling key %s", ev.Name())
		switch ev.Key() {
		case tcell.KeyDown:
			log.Printf("Handling KeyDown")
			err := v.pipeline.ScrollDownLineBuffer()
			if err != nil {
				log.Printf("Eventloop ScrollDownLineBufferError")
				screen.Beep()
				return true
			}
			v.Render(true)
			return true
		case tcell.KeyUp:
			err := v.pipeline.ScrollUpLineBuffer()
			if err != nil {
				log.Printf("Eventloop ScrollUpLineBufferError")
				screen.Beep()
				return true
			}
			v.Render(true)
			return true

		/*case tcell.KeyRight:
		    v.Move(1, 0)
		    return true
		case tcell.KeyLeft:
		    v.Move(-1, 0)
		    return true*/
		case tcell.KeyCtrlF, tcell.KeyPgDn:
			// TODO: would be surprised if this still works
			v.Move(0, v.viewHeight)
			return true
		case tcell.KeyCtrlB, tcell.KeyPgUp:
			v.Move(0, -v.viewHeight)
			return true
		case tcell.KeyCtrlA, tcell.KeyHome:
			v.SetCursor(0, 0)
			v.Render(true)
		case tcell.KeyCtrlE, tcell.KeyEnd:
			_, length, err := v.pipeline.Size()
			if err == nil {
				v.SetCursor(0, length-10)
				v.Render(true)
			}
		}
	case *model.EventFilterOutput:
		log.Println("Handling EventFilterOutput")
		v.Render(true)
	}

	return false
}

func (v *View) SetShowLineNumbers(showLineNumbers bool) {
	v.showLineNumbers = showLineNumbers
}
