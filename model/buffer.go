package model

import (
	"bufio"
	"fmt"
	"log"
	"os"

	//"time"

	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type Buffer struct {
	filePath     string
	width        int
	lines        []Line
	eventHandler tcell.EventHandler
}

func NewBufferFromFile(filePath string) (*Buffer, error) {
	b := Buffer{}
	err := b.readFromFile(filePath)
	if err != nil {
		return &b, err
	}

	return &b, nil
}

func (b *Buffer) Size() (int, int, error) {
	return b.width, len(b.lines), nil
}

func (b *Buffer) readFromFile(filePath string) error {
	//log.Println("Called readLinesFromFile")

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	b.filePath = filePath

	scanner := bufio.NewScanner(file)

	b.lines = nil
	b.width = 0
	y := 0
	for scanner.Scan() {
		text := scanner.Text()
		newLine := Line{y, LineWithoutStatus, text, make([]uint8, len(text))}
		if len(newLine.Str) != len(newLine.ColorIndex) {
			log.Panicf("Line length mismatch: %d != %d", len(newLine.Str), len(newLine.ColorIndex))
		}
		b.lines = append(b.lines, newLine)
		b.width = util.IntMax(len(newLine.Str)-1, b.width)
		y++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return nil
}

func (b *Buffer) GetLine(line int) (Line, error) {
	if line < 0 || line >= len(b.lines) {
		return Line{Str: "ErrOutOfBounds"}, util.ErrOutOfBounds
	}
	b.decolorizeLine(b.lines[line])
	b.lines[line].Status = LineWithoutStatus
	return b.lines[line], nil
}

func (b *Buffer) decolorizeLine(line Line) {
	for i := range line.ColorIndex {
		line.ColorIndex[i] = 0
	}
}

func (b *Buffer) Source() (Filter, error) {
	return nil, fmt.Errorf("buffers don't have a source")
}

func (b *Buffer) SetSource(source Filter) {
	log.Panicln("SetSource() should never be called on a buffer!")
}

func (b *Buffer) Watch(eventHandler tcell.EventHandler) {
	b.eventHandler = eventHandler
}

func (b *Buffer) Unwatch(eventHandler tcell.EventHandler) {
	// TODO definitely fix this
	b.eventHandler = nil
}

func (b *Buffer) HandleEvent(ev tcell.Event) bool {
	return b.eventHandler.HandleEvent(ev)
}

func (b *Buffer) SetFilterFunc(fn func(input string, key string) (string, error)) {
	log.Panicln("SetFilterFunc() should never be called on a buffer!")
}

func (b *Buffer) SetKey(key string) error {
	log.Panicln("SetKey() should never be called on a buffer!")
	return nil
}

func (b *Buffer) SetMode(mode int) {
	log.Panicln("SetMode() should never be called on a buffer!")
}

func (b *Buffer) SetCaseSensitive(caseSensitive bool) error {
	log.Panicln("SetCaseSensitive() should never be called on a buffer!")
	return nil
}

func (b *Buffer) SetColorIndex(colorIndex uint8) {
	// Buffers don't have a color
}
