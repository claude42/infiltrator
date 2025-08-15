package model

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model/reader"
	"github.com/claude42/infiltrator/util"
)

var displayLock sync.Mutex

type Display struct {
	// this is the buffer that should be rendered on screen. It can potentially
	// happen that its size is out of sync with the actual screen size then
	// its data should be ignored until there's an updated version with the
	// correct dimensions
	Buffer []*reader.Line

	// at what percentage of the whole buffer are we currently
	// TODO: decide: display percentag in relation to whole file or to the
	// filters' output. Latter option obviously will be very expensive.
	Percentage int

	// current column to dispaly in the first screen column
	CurrentCol int

	// the following parameters reference the source (or lines in the source) -
	// NOT the screen buffer
	TotalLength  int
	CurrentMatch int
}

// Initializes a 25 line display. Height likely will get overwritten
// immediately but in this way, inital calls refreshDisplay() will not fail.
func NewDisplay() *Display {
	return &Display{
		Buffer:       make([]*reader.Line, 25),
		CurrentMatch: -1,
	}
}

// does not lock!!!
func (d *Display) Height() int {
	return len(d.Buffer)
}

// does lock
func (d *Display) SetHeight(height int) {
	displayLock.Lock()
	defer displayLock.Unlock()

	currentHeight := len(d.Buffer)

	if height == currentHeight {
		return
	} else if height < currentHeight {
		d.Buffer = d.Buffer[:height]
		return
	}

	// so height > currentHeight
	d.Buffer = append(d.Buffer, make([]*reader.Line, height-currentHeight)...)

	var lineNo int
	if currentHeight > 0 && d.Buffer[currentHeight-1] != nil {
		lineNo = d.Buffer[currentHeight-1].No + 1
		if lineNo == 0 {
			d.fillRestOfBufferWithNonExistingLines(currentHeight - 1)
			return
		}
	} else {
		lineNo = 0
	}
	// copied from refreshDisplay()
	y := currentHeight
	for y < height {
		// log.Printf("doing line=%d", y)
		line, err := GetFilterManager().getLine(lineNo)
		lineNo++
		if errors.Is(err, util.ErrOutOfBounds) {
			break
		} else if err != nil {
			log.Panicf("fuck me: %v", err)
		}

		if line.Status != reader.LineHidden {
			d.Buffer[y] = line
			y++
		}
	}

	for ; y < height; y++ {
		d.Buffer[y] = reader.NonExistingLine
	}

	d.Percentage = GetFilterManager().percentage()
}

func (d *Display) fillRestOfBufferWithNonExistingLines(y int) {
	for ; y < len(d.Buffer); y++ {
		d.Buffer[y] = reader.NonExistingLine
	}
}

// lock
func (d *Display) SetTotalLength(length int) {
	displayLock.Lock()
	d.TotalLength = length
	displayLock.Unlock()
}

// does not lock
func (d *Display) UnsetCurrentMatch() {
	d.CurrentMatch = -1
}

func (d *Display) SetCurrentCol(newCurrentCol int) {
	displayLock.Lock()
	d.CurrentCol = newCurrentCol
	displayLock.Unlock()
}

// does lock
// Will return the new startLine - in case the original starting line (or subsequent
// lines) didn't match the filters
func (d *Display) refreshDisplay(ctx context.Context, wg *sync.WaitGroup,
	lineNo int) {

	if wg != nil {
		defer wg.Done()
	}

	displayLock.Lock()
	defer displayLock.Unlock()

	displayHeight := d.Height()
	if displayHeight == 0 {
		return
	}

	y := 0
	for y < displayHeight {
		// log.Printf("doing line=%d", y)
		line, err := GetFilterManager().getLine(lineNo)
		lineNo++
		if errors.Is(err, util.ErrOutOfBounds) {
			break
		} else if err != nil {
			log.Panicf("fuck me: %v", err)
		}

		if line.Status != reader.LineHidden {
			d.Buffer[y] = line
			y++
		}
		if ctx != nil {
			select {
			case <-ctx.Done():
				return
			default:
				// continue
			}
		}
	}

	d.fillRestOfBufferWithNonExistingLines(y)

	d.Percentage = GetFilterManager().percentage()

	config.GetConfiguration().PostEventFunc(NewEventDisplay(*d))
}

func (d *Display) firstLine() *reader.Line {
	return d.Buffer[0]
}

func (d *Display) lastLine() *reader.Line {
	return d.Buffer[len(d.Buffer)-1]
}

func (d *Display) searchOnScreen(startOnScreen int, direction scrollDirection) (*reader.Line, error) {
	height := len(d.Buffer)

	for i := startOnScreen; i >= 0 && i < height; i = i + int(direction) {
		if d.Buffer[i].Matched {
			return d.Buffer[i], nil
		}
	}

	return nil, util.ErrNotFound
}

func (d *Display) isAffectedByNewContend() bool {
	// only if display is currently at the end and not all lines of the
	// display are filled, then new lines in the file will affect the display.
	if len(d.Buffer) == 0 || d.lastLine() == nil {
		return true
	}
	return d.lastLine().No == -1
}

func (d *Display) addLineAtBottomRemoveLineAtTop(line *reader.Line) {
	if d.Height() > 0 {
		d.Buffer = append(d.Buffer[1:], line)
	} else {
		d.Buffer = []*reader.Line{line}
	}
}

func (d *Display) addLineAtTopRemoveLineAtBottom(line *reader.Line) {
	if d.Height() > 0 {
		d.Buffer = append([]*reader.Line{line},
			d.Buffer[:d.Height()-1]...)
	} else {
		d.Buffer = []*reader.Line{line}
	}
}
