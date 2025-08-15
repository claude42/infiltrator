package filter

import (
	"sync"

	"github.com/claude42/infiltrator/model/lines"
)

type Cache struct {
	FilterImpl
	sync.Mutex
	lines map[int]*lines.Line
}

func NewCache() *Cache {
	c := &Cache{}
	c.lines = make(map[int]*lines.Line)

	return c
}

func (c *Cache) GetLine(lineNo int) (*lines.Line, error) {
	c.Lock()
	defer c.Unlock()
	line, ok := c.lines[lineNo]
	if ok {
		return line, nil
	}

	sourceLine, err := c.source.GetLine(lineNo)
	if err != nil {
		return sourceLine, err
	}
	c.lines[lineNo] = sourceLine

	return sourceLine, nil
}

func (c *Cache) Invalidate() {
	c.Lock()
	c.lines = nil
	c.lines = make(map[int]*lines.Line)
	c.Unlock()
}
