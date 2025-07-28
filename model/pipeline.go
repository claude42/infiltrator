package model

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

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

	screen Screen

	currentLine int
}

type Screen struct {
	sync.Mutex

	buffer []Line
	clean  bool
}

func (s *Screen) Length() int {
	return len(s.buffer)
}

type EventPositionChanged struct {
	time time.Time
}

func NewEventPositionChanged() *EventPositionChanged {
	e := &EventPositionChanged{}
	e.time = time.Now()

	return e
}

func (e *EventPositionChanged) When() time.Time {
	return e.time
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
			p.screen.clean = false
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
		p.screen.clean = false
	}

	return p.PostEvent(ev)
}

// Will return the new startLine - in case the original starting line (or subsequent
//
//	lines) didn't match the filters
func (p *Pipeline) RefreshScreenBuffer(startLine, viewHeight int) {
	lineNo := startLine
	y := 0
	p.screen.buffer = make([]Line, viewHeight)
	for y < viewHeight {
		line, err := p.GetLine(lineNo)
		lineNo++
		if errors.Is(err, util.ErrOutOfBounds) {
			break
		} else if err != nil {
			log.Panicf("fuck me: %v", err)
		}

		if line.Status != LineHidden {
			p.screen.buffer[y] = line
			y++
		}
	}

	for ; y < viewHeight; y++ {
		p.screen.buffer[y] = Line{-1, LineDoesNotExist, "", []uint8{}}
	}

	p.screen.clean = true
}

// will return the line it scrolled to
func (p *Pipeline) ScrollDownLineBuffer(updatePosition bool) (Line, error) {
	var nextLine Line
	var err error

	_, length, err := p.Size()
	if err != nil {
		// TODO error handling
		return Line{}, err
	}

	if p.screen.Length() <= 0 {
		// TODO: better error handling
		return Line{}, util.ErrOutOfBounds
	}

	lastLineOnScreen := p.screen.buffer[p.screen.Length()-1]
	if lastLineOnScreen.Status == LineDoesNotExist {
		return Line{}, util.ErrOutOfBounds
	}

	lineNo := lastLineOnScreen.No + 1
	// lineNo := p.screenBuffer[p.screen.Length()-1].No + 1

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
		return Line{}, util.ErrOutOfBounds
	}

	if p.screen.Length() > 0 {
		p.screen.buffer = append(p.screen.buffer[1:], nextLine)
	} else {
		p.screen.buffer = []Line{nextLine}
	}

	p.currentLine = p.screen.buffer[0].No
	if updatePosition {
		p.PostEvent(NewEventPositionChanged())
	}

	return nextLine, nil
}

// will return the line it scrolled to
func (p *Pipeline) ScrollUpLineBuffer() (Line, error) {
	var prevLine Line

	lineNo := p.screen.buffer[0].No - 1

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
		return Line{}, util.ErrOutOfBounds
	}

	if p.screen.Length() > 0 {
		p.screen.buffer = append([]Line{prevLine}, p.screen.buffer[:p.screen.Length()-1]...)
	} else {
		p.screen.buffer = []Line{prevLine}
	}

	p.currentLine = p.screen.buffer[0].No
	p.PostEvent(NewEventPositionChanged())

	return prevLine, nil
}

func (p *Pipeline) FindNextMatch(start int) (int, error) {
	_, length, err := p.Size()
	if err != nil {
		return -1, err
	}

	for i := start; ; i++ {
		newLine, err := p.GetLine(i)
		if err != nil || i >= length {
			return -1, util.ErrOutOfBounds
		} else if newLine.Status == LineMatched {
			return newLine.No, nil
		}
	}
}

func (p *Pipeline) FindPrevMatch(start int) (int, error) {
	for i := start; ; i-- {
		newLine, err := p.GetLine(i)
		if err != nil || i < 0 {
			return -1, util.ErrOutOfBounds
		} else if newLine.Status == LineMatched {
			return newLine.No, nil
		}
	}
}

func (p *Pipeline) ScreenBuffer(startLine, viewHeight int) []Line {
	// TODO: double-check startLine, viewHeight here
	if !p.screen.clean {
		p.RefreshScreenBuffer(startLine, viewHeight)
	}
	return p.screen.buffer
}

func (p *Pipeline) SetCurrentLine(newCurrentLine int) {
	p.currentLine = newCurrentLine
	p.InvalidateScreenBuffer()
	p.PostEvent(NewEventPositionChanged())
}

func (p *Pipeline) InvalidateScreenBuffer() {
	p.screen.clean = false
}

func (p *Pipeline) Percentage() (int, error) {
	_, length, err := p.Size()
	if err != nil || p.currentLine < 0 ||
		p.currentLine > length {
		return -1, err
	}

	percentage := 100 * (p.currentLine + p.screen.Length()) / length
	if percentage > 100 {
		percentage = 100
	}

	return percentage, nil
}

// func (p *Pipeline) FindFilteredEnd() (int, error) {
// 	for {
// 		_, err := p.ScrollDownLineBuffer()
// 		if err != nil {
// 			break
// 		}
// 	}
// }

// func (p *Pipeline) FindFilteredEnd() (int, error) {
// 	_, length, err := p.Size()
// 	if err != nil {
// 		return -1, err
// 	}
// 	viewHeight := p.screen.Length()

// 	// search number of view lines times a line which is not hidden
// 	// starting from the end of the file

// 	var line Line
// 	for lineUp := 0; lineUp < viewHeight-1; lineUp++ {
// 		for i := length - 1; ; i-- {
// 			if i < 0 {
// 				return 0, nil
// 			}
// 			line, err := p.GetLine(i)
// 			if err != nil {
// 				return -1, err
// 			}
// 			status := line.Status
// 			if status == LineWithoutStatus || status == LineMatched ||
// 				status == LineDimmed {
// 				break
// 			} else if i < 0 {
// 				return -1, fmt.Errorf("all lines vanished")
// 			}
// 		}
// 	}

// 	return line.No, nil
// }
