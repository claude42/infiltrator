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
		newLine := Line{y, scanner.Text()}
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
	log.Printf("Buffer.GetLine(%d), len(b.lines)=%d", line, len(b.lines))
	if line < 0 || line >= len(b.lines) {
		log.Println("ErrOutOfBounds")
		return Line{}, ErrOutOfBounds
	}
	return b.lines[line], nil
}

func (b *Buffer) Source() (Filter, error) {
	return nil, fmt.Errorf("buffers don't have a source")
}

func (b *Buffer) SetSource(source Filter) {
	log.Fatalln("SetSource() should never be called on a buffer!")
}

func (b *Buffer) SetEventHandler(eventHandler tcell.EventHandler) {
	log.Println("Buffer.SetEventHandler")
	b.eventHandler = eventHandler
}

func (b *Buffer) HandleEvent(ev tcell.Event) bool {
	return b.eventHandler.HandleEvent(ev)
}
