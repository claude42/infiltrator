package model

import (
	"errors"
	"fmt"
	"log"

	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type Pipeline struct {
	filters      []Filter
	eventHandler tcell.EventHandler

	screenBuffer      []Line
	screenBufferClean bool
}

var instance *Pipeline

func GetPipeline() *Pipeline {
	// not thread safe!
	if instance == nil {
		instance = &Pipeline{}
	}
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

func (p *Pipeline) Watch(eventHandler tcell.EventHandler) {
	p.eventHandler = eventHandler
}

func (p *Pipeline) HandleEvent(ev tcell.Event) bool {
	switch ev.(type) {
	case *EventFilterOutput:
		p.screenBufferClean = false
	}

	if p.eventHandler == nil {
		return false
	}
	return p.eventHandler.HandleEvent(ev)
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
		if errors.Is(err, util.ErrLineDidNotMatch) {
			log.Panicf("shoult not happen anymore")
		} else if errors.Is(err, util.ErrOutOfBounds) {
			break
		} else if err != nil {
			log.Panicf("fuck me")
		}

		if line.Status != LineHidden {
			p.screenBuffer[y] = line
			y++
		}
	}

	for ; y < viewHeight; y++ {
		p.screenBuffer[y] = Line{-1, LineWithoutStatus, "", []uint8{}}
	}

	p.screenBufferClean = true
}

func (p *Pipeline) ScrollDownLineBuffer() error {
	// TODO: get rid of Move(), what about MoveCursor()
	var nextLine Line
	var err error

	_, length, err := p.Size()
	if err != nil {
		// TODO error handling
		return err
	}
	lineNo := p.screenBuffer[len(p.screenBuffer)-1].No + 1
	log.Printf("length=%d, lineNo=%d", length, lineNo)

	for ; lineNo < length; lineNo++ {
		nextLine, err = p.GetLine(lineNo)
		if err == nil {
			// matching line found
			break
		}
	}

	// Could have also checked for err, not sure
	// what's more elegant...
	if lineNo >= length {
		log.Println("P.ScrollDown ErrOutOfBounds")
		return util.ErrOutOfBounds
	}

	log.Println("Pipeline.ScrollDown3")

	if len(p.screenBuffer) > 0 {
		p.screenBuffer = append(p.screenBuffer[1:], nextLine)
	} else {
		p.screenBuffer = []Line{nextLine}
	}
	log.Printf("Pipeline scrolled down one line")

	return nil
}

func (p *Pipeline) ScrollUpLineBuffer() error {
	var prevLine Line
	var err error

	lineNo := p.screenBuffer[0].No - 1

	for ; lineNo >= 0; lineNo-- {
		prevLine, err = p.GetLine(lineNo)
		if err == nil {
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
