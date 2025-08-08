package reader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model/busy"
	"github.com/fsnotify/fsnotify"
)

var (
	readerInstance *Reader
	readerOnce     sync.Once
)

type Reader struct {
}

func GetReader() *Reader {
	readerOnce.Do(func() {
		readerInstance = &Reader{}
	})
	return readerInstance
}

func (r *Reader) ReadFromFile(filePath string, context context.Context,
	ch chan<- []*Line, follow bool) {

	defer config.GetConfiguration().WaitGroup.Done()

	quit := config.GetConfiguration().Quit

	file, err := os.Open(filePath)
	if err != nil {
		quit <- err.Error()
		close(quit)
		return
	}
	defer file.Close()

	lineNo, err := r.readNewLines(file, ch, 0)
	if err != nil {
		quit <- err.Error()
		close(quit)
		return
	}

	if !follow {
		return
	}

	if config.GetConfiguration().FollowFile {
		r.startWatching(filePath, file, context, ch, lineNo)
	}
}

func (r *Reader) ReopenForWatching(filePath string, context context.Context,
	ch chan<- []*Line, lineNo int) {

	log.Println("ReopenForWatching")

	defer config.GetConfiguration().WaitGroup.Done()

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("error opening file %s: %+v", filePath, err)
		return
	}
	defer file.Close()

	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		log.Printf("error seeking in file %s: %+v", filePath, err)
		return
	}

	r.startWatching(filePath, file, context, ch, lineNo)
	log.Println("ReopenForWatching ended")
}

func (r *Reader) startWatching(filePath string, file *os.File,
	context context.Context, ch chan<- []*Line, lineNo int) {

	log.Println("Start watching")
	watcher, err := r.initWatcher(filePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer watcher.Close()

	err = r.keepWatching(watcher, file, context, ch, lineNo)
	if err != nil {
		log.Println(err)
		return
	}
}

func (r *Reader) ReadFromStdin(ch chan<- []*Line) {
	quit := config.GetConfiguration().Quit

	yes, err := r.canUseStdin()
	if err != nil {
		quit <- err.Error()
		close(quit)
		return
	} else if !yes {
		quit <- "Missing filename"
		close(quit)
		return
	}

	lineNo := 0
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		busy.Spin()
		ch <- []*Line{NewLine(lineNo, text)}
		lineNo++
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading file: %+v", err)
	}
}

func (r *Reader) canUseStdin() (bool, error) {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}

	if (fileInfo.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		return false, nil
	}

	return true, nil
}

func (r *Reader) keepWatching(watcher *fsnotify.Watcher, file *os.File,
	context context.Context, ch chan<- []*Line, lineNo int) error {
	var err error
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Watcher errors channel closed.")
				return nil
			}

			if event.Has(fsnotify.Write) {
				lineNo, err = r.readNewLines(file, ch, lineNo)
				if err != nil {
					return fmt.Errorf("error reading file, %w", err)
				}

				continue
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("Watcher errors channel closed.")
				return nil
			}
			log.Printf("Watcher error: %+v", err)
			continue
		case <-context.Done():
			log.Println("Reader received shutdown")
			return nil
		}
	}
}

func (r *Reader) initWatcher(filePath string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating watcher: %w", err)
	}

	err = watcher.Add(filePath)
	if err != nil {
		return nil, fmt.Errorf("error watching file %s: %w", filePath, err)
	}
	return watcher, nil
}

func (r *Reader) readNewLines(file *os.File, ch chan<- []*Line, lineNo int) (int, error) {
	var newLines []*Line

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		busy.Spin()
		newLines = append(newLines, NewLine(lineNo, text))
		lineNo++
	}

	if err := scanner.Err(); err != nil {
		return lineNo, fmt.Errorf("error reading file: %w", err)
	}

	ch <- newLines

	return lineNo, nil
}
