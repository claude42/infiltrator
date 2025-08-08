package model

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

func (c *Cache) getLine(lineNo int) (*reader.Line, error) {
	c.Lock()
	defer c.Unlock()
	line, ok := c.lines[lineNo]
	if ok {
		return line, nil
	}

	sourceLine, err := c.source.getLine(lineNo)
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

func (c *Cache) setSource(source Filter) {
	c.source = source
}

func (c *Cache) size() (int, int) {
	return c.source.size()
}

func (c *Cache) length() int {
	return c.source.length()
}

func (c *Cache) setKey(key string) error {
	log.Panicln("SetKey() should never be called on a cache!")
	return nil
}

func (c *Cache) setMode(mode FilterMode) {
	log.Panicln("SetMode() should never be called on a cache!")
}

func (c *Cache) setCaseSensitive(caseSensitive bool) error {
	log.Panicln("SetCaseSensitive() should never be called on a cache!")
	return nil
}

func (c *Cache) setColorIndex(colorIndex uint8) {
	log.Panicln("setColorIndex() should never be called on a cache!")
}
