package reader

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model/busy"
	"github.com/claude42/infiltrator/model/formats"
	"github.com/claude42/infiltrator/model/lines"
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

func (r *Reader) ReadFromFile(ctx context.Context, wg *sync.WaitGroup,
	quit chan<- string, filePath string, ch chan<- []*lines.Line, follow bool) {

	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		quit <- err.Error()
		close(quit)
		return
	}
	defer file.Close()

	isGzip, _ := formats.IsGzip(file)

	var ioReader io.Reader
	if isGzip {
		gzReader, err := gzip.NewReader(file)
		fail.OnError(err, "Failed to create gzip reader")
		defer gzReader.Close()
		ioReader = gzReader
	} else {
		ioReader = file
	}

	lineNo, err := r.readNewLines(ioReader, ch, 0)
	if err != nil {
		quit <- err.Error()
		close(quit)
		return
	}

	if !config.GetConfiguration().FollowFile {
		return
	}

	r.startWatching(ctx, filePath, file, ch, lineNo)

}

func (r *Reader) ReopenForWatching(ctx context.Context, wg *sync.WaitGroup,
	filePath string, ch chan<- []*lines.Line, lineNo int) {

	log.Println("ReopenForWatching")

	defer wg.Done()

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

	r.startWatching(ctx, filePath, file, ch, lineNo)
	log.Println("ReopenForWatching ended")
}

func (r *Reader) startWatching(ctx context.Context, filePath string,
	file *os.File, ch chan<- []*lines.Line, lineNo int) {

	log.Println("Start watching")
	watcher, err := r.initWatcher(filePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer watcher.Close()

	err = r.keepWatching(ctx, watcher, file, ch, lineNo)
	if err != nil {
		log.Println(err)
		return
	}
}

func (r *Reader) ReadFromStdin(ch chan<- []*lines.Line, quit chan<- string) {

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
		ch <- []*lines.Line{lines.NewLine(lineNo, text)}
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

func (r *Reader) keepWatching(ctx context.Context, watcher *fsnotify.Watcher,
	file *os.File, ch chan<- []*lines.Line, lineNo int) error {
	var err error
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Watcher events channel closed.")
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
		case <-ctx.Done():
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

func (r *Reader) readNewLines(file io.Reader, ch chan<- []*lines.Line, lineNo int) (int, error) {
	var newLines []*lines.Line

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		busy.Spin()
		newLines = append(newLines, lines.NewLine(lineNo, text))
		lineNo++
	}

	if err := scanner.Err(); err != nil {
		return lineNo, fmt.Errorf("error reading file: %w", err)
	}

	ch <- newLines

	return lineNo, nil
}
