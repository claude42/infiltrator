package model

import (
	"errors"
	"fmt"
	"log"

	// "github.com/claude42/infiltrator/util"

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
	log.Printf("Pipeline.AddFilter")
	var last Filter

	// add pipeline as eventhandler of the new filter
	f.SetEventHandler(p)

	if len(p.filters) > 0 {
		log.Printf("Adding to existing filters")
		// remove pipeline itself as the event handler of previous
		// filter, instead add new filter as event handler
		last = p.filters[len(p.filters)-1]
		last.SetEventHandler(f)
		f.SetSource(last)
	} else {
		log.Printf("No other filters yet")
	}
	p.filters = append(p.filters, f)
}

func (p *Pipeline) GetOutputFilter() (Filter, error) {
	if len(p.filters) == 0 {
		return nil, fmt.Errorf("pipeline empty")
	}

	return p.filters[len(p.filters)-1], nil
}

func (p *Pipeline) GetLine(line int) (Line, error) {
	filter, err := p.GetOutputFilter()
	if err != nil {
		return Line{}, err
	}

	readLine, err := filter.GetLine(line)
	log.Printf("P.Getline %s", readLine.Str)
	return readLine, err
}

func (p *Pipeline) Size() (int, int, error) {
	filter, err := p.GetOutputFilter()
	if err != nil {
		return 0, 0, err
	}

	return filter.Size()
}

func (p *Pipeline) SetEventHandler(eventHandler tcell.EventHandler) {
	log.Printf("Pipeline.SetEventHandler()")
	p.eventHandler = eventHandler
	// has that ever been a good idea?
	// p.HandleEvent(NewEventFilterOutput())
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
	log.Printf("Pipieline.refresh")
	lineNo := startLine
	y := 0
	p.screenBuffer = make([]Line, viewHeight)
	for y < viewHeight {
		line, err := p.GetLine(lineNo)
		lineNo++
		if errors.Is(err, ErrLineDidNotMatch) {
			log.Printf("Pipeline.refresh ErrLineDidNotMatch")
			continue
		} else if errors.Is(err, ErrOutOfBounds) {
			log.Printf("Pipeline.refresh ErrOutOfBounds")
			break
		} else if err != nil {
			log.Fatalf("fuck me")
		}
		p.screenBuffer[y] = line
		y++
	}

	for ; y < viewHeight; y++ {
		p.screenBuffer[y] = Line{-1, ""}
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
		return ErrOutOfBounds
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
		return ErrOutOfBounds
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
