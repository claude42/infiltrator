package ui

import (
	"fmt"
	"log"

	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type View struct {
	viewWidth, viewHeight int
	curX, curY            int
	showLineNumbers       bool
	followFile            bool
	currentMatchLineNo    int

	ComponentImpl
}

func NewView() *View {
	v := &View{}
	v.viewWidth, v.viewHeight = screen.Size()
	model.GetPipeline().Watch(v)
	v.SetCursor(0, 0)
	v.currentMatchLineNo = -1

	return v
}

func (v *View) Render(updateScreen bool) {
	screenBuffer := model.GetPipeline().ScreenBuffer(v.curY, v.viewHeight)
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
	matched := line.No == v.currentMatchLineNo

	if v.showLineNumbers {
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
	_, length, err := model.GetPipeline().Size()
	if err != nil {
		return 0 // TODO: 0 ok?
	}

	str := fmt.Sprintf("%*d ", util.CountDigits(length-1), line.No)

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

func (v *View) SetCursor(x, y int) error {
	var err error

	if v.curY != y {
		model.GetPipeline().InvalidateScreenBuffer()
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

	width, length, err := model.GetPipeline().Size()
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
	model.GetPipeline().RefreshScreenBuffer(v.curY, v.viewHeight)
}

func (v *View) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *model.EventBufferDirty:
		v.reactToFileUpdate()
		return true
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case '<':
				v.scrollHome()
				return true
			case '>', 'G':
				v.scrollEnd()
				return true
			case ' ':
				v.pageDown()
				return true
			case 'j':
				v.scrollDown()
				return true
			case 'k':
				v.scrollUp()
				return true
			case 'h':
				v.scrollHorizontal(-1)
				return true
			case 'l':
				v.scrollHorizontal(1)
				return true
			case 'n':
				v.goToNextMatchingLine()
				return true
			case 'N':
				v.goToPrevMatchingLine()
			}
		case tcell.KeyDown, tcell.KeyEnter:
			v.scrollDown()
			return true
		case tcell.KeyUp:
			v.scrollUp()
			return true
		case tcell.KeyRight:
			v.scrollHorizontal(1)
			return true
		case tcell.KeyLeft:
			v.scrollHorizontal(-1)
			return true
		case tcell.KeyCtrlF, tcell.KeyPgDn:
			v.pageDown()
			return true
		case tcell.KeyCtrlB, tcell.KeyPgUp:
			v.pageUp()
			return true
		case tcell.KeyCtrlA, tcell.KeyHome:
			v.scrollHome()
			return true
		case tcell.KeyCtrlE, tcell.KeyEnd:
			v.scrollEnd()
			return true
		}
	case *tcell.EventMouse:
		buttons := ev.Buttons()
		log.Printf("Wheel: %d", buttons)

		// Horizontal mouse wheel doesn't seem to work with the terminals I
		// have access to but we'll leave it in anyways...
		if buttons&tcell.WheelUp != 0 {
			v.scrollUp()
			return true
		} else if buttons&tcell.WheelDown != 0 {
			v.scrollDown()
			return true
		} else if buttons&tcell.WheelLeft != 0 {
			v.scrollHorizontal(-1)
			return true
		} else if buttons&tcell.WheelRight != 0 {
			v.scrollHorizontal(1)
			return true
		}
	case *model.EventFilterOutput:
		v.currentMatchLineNo = -1
		v.Render(true)
	}

	return false
}

func (v *View) goToPrevMatchingLine() {
	var startSearchWith int
	if v.isOnScreen(v.currentMatchLineNo) {
		startSearchWith = v.currentMatchLineNo - 1
	} else {
		startSearchWith = v.curY + v.viewHeight - 1
	}

	newLineNo, err := model.GetPipeline().FindPrevMatch(startSearchWith)
	if err != nil {
		screen.Beep()
		return
	}

	if v.isOnScreen(newLineNo) {
		v.currentMatchLineNo = newLineNo
		v.Render(true)
		return
	}

	// TODO: this is probably not very efficient, espcially as we have
	// determined the matching line number already above.
	// Pipeline should provide some ScrollUpTo() and ScrollDownTo() methods

	for {
		newLine, err := model.GetPipeline().ScrollUpLineBuffer()
		if err != nil {
			screen.Beep()
			v.Render(true)
			return
		} else if newLine.Status == model.LineMatched {
			v.currentMatchLineNo = newLine.No
			break
		}
	}

	// scroll a bit further to see a bit more context around the highlighted match

	for i := 0; i < v.viewHeight/4; i++ {
		_, err := model.GetPipeline().ScrollUpLineBuffer()
		if err != nil {
			break
		}
	}

	v.Render(true)
}

func (v *View) goToNextMatchingLine() {
	var startSearchWith int
	if v.isOnScreen(v.currentMatchLineNo) {
		startSearchWith = v.currentMatchLineNo + 1
	} else {
		startSearchWith = v.curY
	}

	newLineNo, err := model.GetPipeline().FindNextMatch(startSearchWith)
	if err != nil {
		screen.Beep()
		return
	}

	if v.isOnScreen(newLineNo) {
		v.currentMatchLineNo = newLineNo
		v.Render(true)
		return
	}

	for {
		newLine, err := model.GetPipeline().ScrollDownLineBuffer()
		if err != nil {
			screen.Beep()
			v.Render(true)
			return
		} else if newLine.Status == model.LineMatched {
			v.currentMatchLineNo = newLine.No
			break
		}
	}

	// scroll a bit further to see a bit more context around the highlighted match

	for i := 0; i < v.viewHeight/4; i++ {
		_, err := model.GetPipeline().ScrollDownLineBuffer()
		if err != nil {
			break
		}
	}

	v.Render(true)
}

func (v *View) isOnScreen(lineNo int) bool {
	if lineNo == -1 {
		return false
	}

	sb := model.GetPipeline().ScreenBuffer(v.curY, v.viewHeight)

	for _, line := range sb {
		if lineNo == line.No {
			return true
		}
	}

	return false
}

func (v *View) reactToFileUpdate() {
	model.GetPipeline().InvalidateScreenBuffer()
	if v.followFile {
		v.scrollEnd()
	} else {
		v.Render(true)
	}
}

func (v *View) scrollUp() {
	_, err := model.GetPipeline().ScrollUpLineBuffer()
	if err != nil {
		screen.Beep()
		return
	}

	v.Render(true)
}

func (v *View) pageUp() {
	var err error
	for i := 0; i < v.viewHeight-1; i++ {
		_, err = model.GetPipeline().ScrollUpLineBuffer()
		if err != nil {
			break
		}
	}
	v.Render(true)
	if err != nil {
		screen.Beep()
	}
}

func (v *View) scrollDown() {
	_, err := model.GetPipeline().ScrollDownLineBuffer()
	if err != nil {
		screen.Beep()
		return
	}

	v.Render(true)
}

func (v *View) pageDown() {
	var err error
	for i := 0; i < v.viewHeight-1; i++ {
		_, err = model.GetPipeline().ScrollDownLineBuffer()
		if err != nil {
			break
		}
	}
	v.Render(true)
	if err != nil {
		screen.Beep()
	}
}

func (v *View) scrollHorizontal(offset int) {
	width, _, err := model.GetPipeline().Size()
	if err != nil {
		// TODO: rather fail than beep?
		screen.Beep()
		return
	}

	newX, err := util.InBetween(v.curX+offset, 0, width)
	if err != nil {
		screen.Beep()
		return
	}

	v.curX = newX
	v.Render(true)
}

func (v *View) scrollHome() {
	v.SetCursor(v.curX, 0)
	v.Render(true)
}

func (v *View) scrollEnd() {
	_, length, err := model.GetPipeline().Size()
	if err != nil {
		screen.Beep()
		return
	}

	v.SetCursor(v.curX, length-v.viewHeight)
	v.Render(true)
}

func (v *View) SetShowLineNumbers(showLineNumbers bool) {
	v.showLineNumbers = showLineNumbers
}

func (v *View) SetFollowFile(followFile bool) {
	v.followFile = followFile
	if followFile {
		v.scrollEnd()
	}
}
