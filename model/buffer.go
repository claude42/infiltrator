package model

import (
	"bufio"
	"fmt"
	"sync"

	// "io"
	"log"
	"os"
	"time"

	//"time"

	"github.com/claude42/infiltrator/util"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	// "github.com/gdamore/tcell/v2"
)

type Buffer struct {
	util.ObservableImpl
	util.EventHandlerIgnoreImpl

	sync.Mutex

	width int
	lines []Line
}

type EventBufferDirty struct {
	time time.Time
}

func NewEventBufferDirty() *EventBufferDirty {
	e := &EventBufferDirty{}
	e.time = time.Now()

	return e
}

func (e *EventBufferDirty) When() time.Time {
	return e.time
}

func NewBuffer() *Buffer {
	return &Buffer{}
}

func (b *Buffer) Size() (int, int, error) {
	b.Lock()
	length := len(b.lines)
	b.Unlock()

	return b.width, length, nil
}

func (b *Buffer) initLines() {
	b.Lock()
	b.lines = nil
	b.Unlock()

	b.width = 0
}

func (b *Buffer) ReadFromFile(filePath string, postEvent func(ev tcell.Event) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	b.initLines()

	lineNo, err := b.addNewLines(file, 0)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(filePath)
	if err != nil {
		return fmt.Errorf("error watching file %s: %s", filePath, err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Watcher events channel closed")
				return nil
			}
			log.Printf("Event received: %s (Op: %s)", event.Name, event.Op.String())

			if event.Has(fsnotify.Write) {
				log.Printf("Handling write")
				lineNo, err = b.addNewLines(file, lineNo)
				if err != nil {
					log.Printf("error reading file %s, %v", filePath, err)
				}

				postEvent(NewEventBufferDirty())

				continue
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("Watcher errors channel closd.")
				return nil
			}
			log.Println("Watcher error:", err)
			continue
		}
	}
}

func (b *Buffer) addNewLines(file *os.File, lineNo int) (int, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		b.addNewLine(lineNo, scanner.Text())
		lineNo++
	}

	if err := scanner.Err(); err != nil {
		return lineNo, fmt.Errorf("error reading file: %w", err)
	}

	return lineNo, nil
}

func (b *Buffer) addNewLine(lineNo int, text string) {
	newLine := Line{lineNo, LineWithoutStatus, text, make([]uint8, len(text))}

	b.Lock()
	b.lines = append(b.lines, newLine)
	b.Unlock()

	b.width = util.IntMax(len(newLine.Str)-1, b.width)
}

func (b *Buffer) GetLine(line int) (Line, error) {
	b.Lock()
	length := len(b.lines)
	b.Unlock()

	if line < 0 || line >= length {
		return Line{Str: "ErrOutOfBounds"}, util.ErrOutOfBounds
	}

	b.Lock()
	b.lines[line].CleanUp()
	b.Unlock()

	return b.lines[line], nil
}

func (b *Buffer) Source() (Filter, error) {
	log.Panicln("Source() should never be called on a buffer")
	return nil, fmt.Errorf("buffers don't have a source")
}

func (b *Buffer) SetSource(source Filter) {
	log.Panicln("SetSource() should never be called on a buffer!")
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
