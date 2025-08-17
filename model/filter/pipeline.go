package filter

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model/busy"
	"github.com/claude42/infiltrator/model/lines"
	"github.com/claude42/infiltrator/util"
)

var pipelineMutex sync.Mutex
var ErrNotEnoughPanels = errors.New("at least one buffer and one filter required")

type Pipeline []Filter

type ScrollDirection int

const (
	DirectionUp   ScrollDirection = -1
	DirectionDown ScrollDirection = 1
)

func (pp *Pipeline) Add(f Filter) {
	pipelineMutex.Lock()
	defer pipelineMutex.Unlock()

	if len(*pp) == 0 {
		*pp = append(*pp, f)
		return
	}

	var pos int
	for pos = len(*pp) - 1; pos > 0; pos-- {
		existingFilter := (*pp)[pos]
		if _, ok := existingFilter.(*Cache); ok {
			continue
		} else if _, ok := existingFilter.(*Source); ok {
			log.Panic("really?!?")
		} else {
			break
		}
	}

	// if the whole for loop went through then pos should be 0 now. So new
	// filter will be added add pos 1, right after the source.

	f.SetSource((*pp)[pos])
	if pos < len(*pp)-1 {
		(*pp)[pos+1].SetSource(f)
		// insert right after fm.filters[pos]
		*pp = append((*pp)[:pos+2], (*pp)[pos+1:]...)
		(*pp)[pos+1] = f
	} else {
		*pp = append(*pp, f)
	}
}

func (pp *Pipeline) Remove(f Filter) error {
	pipelineMutex.Lock()
	defer pipelineMutex.Unlock()
	if len(*pp) <= 2 {
		return ErrNotEnoughPanels
	}
	for i, filter := range *pp {
		if filter == f {
			fail.If(i == 0, "Cannot remove source from pipeline")
			*pp = append((*pp)[:i], (*pp)[i+1:]...)
			if i < len(*pp) {
				(*pp)[i].SetSource((*pp)[i-1])
			}
			return nil
		}
	}
	log.Panicln("Filter not found in pipeline")
	return fmt.Errorf("Filter not found in pipeline")
}

func (pp *Pipeline) GetLine(line int) (*lines.Line, error) {
	filter, err := pp.OutputFilter()
	if err != nil {
		return lines.NonExistingLine, err
	}

	busy.SpinWithFraction(line, pp.SourceLength())

	return filter.GetLine(line)
}

func (pp *Pipeline) Search(start int,
	direction ScrollDirection) (*lines.Line, error) {

	length := pp.SourceLength()

	for i := start; ; i = i + int(direction) {
		busy.SpinWithFraction(i, pp.SourceLength())
		newLine, err := pp.GetLine(i)
		if err != nil || i < 0 || i >= length {
			return nil, util.ErrNotFound
		} else if newLine.Matched {
			return newLine, nil
		}
	}
}

func (pp *Pipeline) FindNonHiddenLine(lineNo int,
	direction ScrollDirection) (*lines.Line, error) {

	fail.If(direction != -1 && direction != 1, "Unknown directionn %d", direction)

	length := pp.SourceLength()

	for lineNo = lineNo + int(direction); lineNo >= 0 && lineNo < length; lineNo = lineNo + int(direction) {
		busy.SpinWithFraction(lineNo, pp.SourceLength())
		prevLine, err := pp.GetLine(lineNo)
		if err != nil {
			return nil, err
		}
		if prevLine.Status == lines.LineWithoutStatus ||
			prevLine.Status == lines.LineMatched ||
			prevLine.Status == lines.LineDimmed {
			return prevLine, nil
		}
	}

	return nil, util.ErrOutOfBounds
}

func (pp *Pipeline) Source() *Source {
	pipelineMutex.Lock()
	fail.If(len(*pp) < 1, "No source in filter stack!")
	firstFilter := (*pp)[0]
	pipelineMutex.Unlock()

	source, ok := firstFilter.(*Source)
	fail.If(!ok, "First filter is not a Source!")
	return source
}

func (pp *Pipeline) OutputFilter() (Filter, error) {
	pipelineMutex.Lock()
	defer pipelineMutex.Unlock()
	if len(*pp) == 0 {
		return nil, fmt.Errorf("pipeline empty")
	}

	return (*pp)[len((*pp))-1], nil
}

func (pp *Pipeline) DateFilter() (*DateFilter, error) {
	pipelineMutex.Lock()
	defer pipelineMutex.Unlock()
	for _, f := range *pp {
		dateFilter, ok := f.(*DateFilter)
		if ok {
			return dateFilter, nil
		}
	}

	return nil, util.ErrNotFound
}

func (pp *Pipeline) Size() (int, int) {
	filter, err := pp.OutputFilter()
	if err != nil {
		return 0, 0
	}

	return filter.Size()
}

func (pp *Pipeline) SourceLength() int {
	filter, err := pp.OutputFilter()
	if err != nil {
		return 0
	}

	return filter.Length()
}

func (pp *Pipeline) InvalidateCaches() {
	for _, f := range *pp {
		cache, ok := f.(*Cache)
		if ok {
			cache.Invalidate()
		}
	}
}
