package model

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

var (
	instance *Pipeline
	once     sync.Once
)

type Pipeline struct {
	util.ObservableImpl

	filters []Filter

	screenBuffer      []Line
	screenBufferClean bool
}

func GetPipeline() *Pipeline {
	once.Do(func() {
		instance = &Pipeline{}
	})
	return instance
}

func (p *Pipeline) AddFilter(f Filter) {
	var last Filter

	// add pipeline as eventhandler of the new filter
	f.Watch(p)

	if len(p.filters) > 0 {
		// remove pipeline itself as the event handler of previous
		// filter, instead add new filter as event handler
		last = p.filters[len(p.filters)-1]
		f.SetSource(last)
	}
	p.filters = append(p.filters, f)
}

func (p *Pipeline) RemoveFilter(f Filter) error {
	if len(p.filters) <= 2 {
		return fmt.Errorf("at least one buffer and one filter required")
	}
	for i, filter := range p.filters {
		if filter == f {
			if i == 0 {
				log.Panicln("Cannot remove buffer from pipeline")
			}
			p.filters = append(p.filters[:i], p.filters[i+1:]...)
			if i < len(p.filters) {
				p.filters[i].SetSource(p.filters[i-1])
			} else {
				// set pipeline as event handler of last filter
				p.filters[i-1].Watch(p)
			}
			f.Unwatch(p)
			p.screenBufferClean = false
			return nil
		}
	}
	log.Panicln("Filter not found in pipeline")
	return fmt.Errorf("Filter not found in pipeline")
}

func (p *Pipeline) OutputFilter() (Filter, error) {
	if len(p.filters) == 0 {
		return nil, fmt.Errorf("pipeline empty")
	}

	return p.filters[len(p.filters)-1], nil
}

func (p *Pipeline) GetLine(line int) (Line, error) {
	filter, err := p.OutputFilter()
	if err != nil {
		return Line{}, err
	}

	return filter.GetLine(line)
}

func (p *Pipeline) Size() (int, int, error) {
	filter, err := p.OutputFilter()
	if err != nil {
		return 0, 0, err
	}

	return filter.Size()
}

func (p *Pipeline) HandleEvent(ev tcell.Event) bool {
	switch ev.(type) {
	case *EventFilterOutput:
		p.screenBufferClean = false
	}

	p.PostEvent(ev)
	return true
}

// Will return the new startLine - in case the original starting line (or subsequent
//
//	lines) didn't match the filters
func (p *Pipeline) RefreshScreenBuffer(startLine, viewHeight int) {
	lineNo := startLine
	y := 0
	p.screenBuffer = make([]Line, viewHeight)
	for y < viewHeight {
		line, err := p.GetLine(lineNo)
		lineNo++
		if errors.Is(err, util.ErrOutOfBounds) {
			break
		} else if err != nil {
			log.Panicf("fuck me: %v", err)
		}

		if line.Status != LineHidden {
			p.screenBuffer[y] = line
			y++
		}
	}

	for ; y < viewHeight; y++ {
		p.screenBuffer[y] = Line{-1, LineDoesNotExist, "", []uint8{}}
	}

	p.screenBufferClean = true
}

func (p *Pipeline) ScrollDownLineBuffer() error {
	var nextLine Line
	var err error

	_, length, err := p.Size()
	if err != nil {
		// TODO error handling
		return err
	}

	lastLineOnScreen := p.screenBuffer[len(p.screenBuffer)-1]
	if lastLineOnScreen.Status == LineDoesNotExist {
		return util.ErrOutOfBounds
	}

	lineNo := lastLineOnScreen.No + 1
	// lineNo := p.screenBuffer[len(p.screenBuffer)-1].No + 1

	for ; lineNo < length; lineNo++ {
		nextLine, _ = p.GetLine(lineNo)
		if nextLine.Status == LineWithoutStatus ||
			nextLine.Status == LineMatched || nextLine.Status == LineDimmed {

			break
		}
	}

	// Could have also checked for err, not sure
	// what's more elegant...
	if lineNo >= length {
		log.Println("P.ScrollDown ErrOutOfBounds")
		return util.ErrOutOfBounds
	}

	if len(p.screenBuffer) > 0 {
		p.screenBuffer = append(p.screenBuffer[1:], nextLine)
	} else {
		p.screenBuffer = []Line{nextLine}
	}

	return nil
}

func (p *Pipeline) ScrollUpLineBuffer() error {
	var prevLine Line

	lineNo := p.screenBuffer[0].No - 1

	for ; lineNo >= 0; lineNo-- {
		prevLine, _ = p.GetLine(lineNo)
		if prevLine.Status == LineWithoutStatus ||
			prevLine.Status == LineMatched || prevLine.Status == LineDimmed {

			// matching line found
			break
		}
	}

	// Could have also checked for err, not sure
	// what's more elegant...
	if lineNo < 0 {
		log.Println("P.ScrollUp ErrOutOfBounds")
		return util.ErrOutOfBounds
	}

	if len(p.screenBuffer) > 0 {
		p.screenBuffer = append([]Line{prevLine}, p.screenBuffer[:len(p.screenBuffer)-1]...)
	} else {
		p.screenBuffer = []Line{prevLine}
	}

	return nil
}

func (p *Pipeline) ScreenBuffer(startLine, viewHeight int) []Line {
	// TODO: double-check startLine, viewHeight here
	if !p.screenBufferClean {
		p.RefreshScreenBuffer(startLine, viewHeight)
	}
	return p.screenBuffer
}

func (p *Pipeline) InvalidateScreenBuffer() {
	p.screenBufferClean = false
}
