package model

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/util"
)

var displayLock sync.Mutex

type Display struct {
	// this is the buffer that should be rendered on screen. It can potentially
	// happen that its size is out of sync with the actual screen size then
	// its data should be ignored until there's an updated version with the
	// correct dimensions
	Buffer []*Line

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

// does not lock!!!
func (d *Display) Height() int {
	return len(d.Buffer)
}

// does lock
func (d *Display) SetHeight(height int) {
	displayLock.Lock()
	defer displayLock.Unlock()
	currentHeight := len(d.Buffer)
	if height < currentHeight {
		d.Buffer = d.Buffer[:height]
	} else if height > currentHeight {
		d.Buffer = append(d.Buffer, make([]*Line, height-currentHeight)...)
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
	lineNo int, unsetCurrentMatch bool) {

	if wg != nil {
		defer wg.Done()
	}

	displayLock.Lock()
	defer displayLock.Unlock()

	if unsetCurrentMatch {
		d.UnsetCurrentMatch()
	}

	displayHeight := d.Height()
	if displayHeight == 0 {
		return
	}

	y := 0
	for y < displayHeight {
		// log.Printf("doing line=%d", y)
		line, err := GetFilterManager().GetLine(lineNo)
		lineNo++
		if errors.Is(err, util.ErrOutOfBounds) {
			break
		} else if err != nil {
			log.Panicf("fuck me: %v", err)
		}

		if line.Status != LineHidden {
			d.Buffer[y] = line
			y++
		}
		if ctx != nil {
			select {
			case <-ctx.Done():
				log.Println("cancelled")
				return
			default:
				// continue
			}
		}
	}

	for ; y < displayHeight; y++ {
		d.Buffer[y] = &Line{-1, LineDoesNotExist, false, "", []uint8{}}
	}

	d.Percentage = GetFilterManager().percentage()

	config.GetConfiguration().PostEventFunc(NewEventDisplay(*d))
	log.Println("went through")
}
