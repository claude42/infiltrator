package filter

import (
	"log"
	"sync"

	"github.com/claude42/infiltrator/model/reader"
)

type Cache struct {
	sync.Mutex
	lines  map[int]*reader.Line
	source Filter
}

func NewCache() *Cache {
	c := &Cache{}
	c.lines = make(map[int]*reader.Line)

	return c
}

func (c *Cache) GetLine(lineNo int) (*reader.Line, error) {
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
	c.lines = make(map[int]*reader.Line)
	c.Unlock()
}

func (c *Cache) SetSource(source Filter) {
	c.source = source
}

func (c *Cache) Size() (int, int) {
	return c.source.Size()
}

func (c *Cache) Length() int {
	return c.source.Length()
}

func (c *Cache) SetKey(key string) error {
	log.Panicln("SetKey() should never be called on a cache!")
	return nil
}

func (c *Cache) SetMode(mode FilterMode) {
	log.Panicln("SetMode() should never be called on a cache!")
}

func (c *Cache) SetCaseSensitive(caseSensitive bool) error {
	log.Panicln("SetCaseSensitive() should never be called on a cache!")
	return nil
}

func (c *Cache) SetColorIndex(colorIndex uint8) {
	log.Panicln("setColorIndex() should never be called on a cache!")
}
