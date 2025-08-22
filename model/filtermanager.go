package model

import (
	"context"
	"log"
	"runtime/debug"
	"sync"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model/busy"
	"github.com/claude42/infiltrator/model/filter"
	"github.com/claude42/infiltrator/model/formats"
	"github.com/claude42/infiltrator/model/lines"
	"github.com/claude42/infiltrator/model/reader"
	"github.com/claude42/infiltrator/util"
)

var (
	filterManagerInstance *FilterManager
	identifyFileTypeOnce  sync.Once
)

type FilterManager struct {
	util.ObservableImpl

	ctx  context.Context
	wg   *sync.WaitGroup
	quit chan<- string

	readerCancelFunc context.CancelFunc

	refresherCancelFunc context.CancelFunc
	refresherWg         sync.WaitGroup

	contentUpdate  chan []*lines.Line
	commandChannel chan Command

	filters     filter.Pipeline
	currentLine int

	display *Display
}

func GetFilterManager() *FilterManager {
	fail.IfNil(filterManagerInstance, "Filtermanager missing!")

	return filterManagerInstance
}

func NewFilterManager(ctx context.Context, wg *sync.WaitGroup, quit chan<- string) *FilterManager {
	fm := &FilterManager{
		ctx:            ctx,
		wg:             wg,
		quit:           quit,
		display:        NewDisplay(),
		contentUpdate:  make(chan []*lines.Line, 10),
		commandChannel: make(chan Command, 10),
	}

	fm.filters.Add(filter.NewSource())
	fm.filters.Add(filter.NewCache())

	filterManagerInstance = fm
	return fm
}

func (fm *FilterManager) ReadFromFile(filePath string) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("A panic occurred: %v\nStack trace:\n%s", r, debug.Stack())
			panic(r)
		}
	}()

	var readCtx context.Context
	readCtx, fm.readerCancelFunc = context.WithCancel(fm.ctx)
	fm.wg.Add(1)
	go reader.GetReader().ReadFromFile(readCtx, fm.wg, fm.quit, filePath,
		fm.contentUpdate, config.UserCfg().Follow)
	// GetLoremIpsumReader().Read(fm.contentUpdate)
}

func (fm *FilterManager) ReadFromStdin() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("A panic occurred: %v\nStack trace:\n%s", r, debug.Stack())
			panic(r)
		}
	}()

	// do NOT add to wait group as the Go routine most likely will not
	// return
	go reader.GetReader().ReadFromStdin(fm.contentUpdate, fm.quit)
	// GetLoremIpsumReader().Read(fm.contentUpdate)
}

func (fm *FilterManager) EventLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("A panic occurred: %v\nStack trace:\n%s", r, debug.Stack())
			panic(r)
		}
	}()

	defer fm.wg.Done()

	for {
		select {
		case newLines := <-fm.contentUpdate:
			log.Printf("Received contentupdate, lines %d-%d", newLines[0].No, newLines[len(newLines)-1].No)
			fm.processContentUpdate(newLines)
		case command := <-fm.commandChannel:
			log.Printf("Received Command: %T", command)
			fm.processCommand(command)
		case <-fm.ctx.Done():
			log.Println("Received shutdown")
			return
		}
	}
}

func (fm *FilterManager) processContentUpdate(newLines []*lines.Line) {
	identifyFileTypeOnce.Do(func() {
		go formats.Identify(newLines)
	})

	// If we're in Follow mode we'll automatically jump to the new end of the
	// file - but only in case we're already at the end
	goToEnd := false
	if config.UserCfg().Follow && fm.alreadyAtTheEnd() {
		goToEnd = true
	}

	length := fm.filters.Source().StoreNewLines(newLines)
	fm.display.SetTotalLength(length)

	// refresh display as necessary
	if goToEnd {
		// two times refreshDisplay() currently necessary because
		// internalScrollEnd() depends on a correctly set up display.
		// UPD: I think this should not be necessary anymore, let's still keep
		// the comment here
		// fm.refreshDisplay()
		// fm.internalScrollEnd()
		fm.internalTail()
	} else if fm.display.isAffectedByNewContend() {
		fm.syncRefreshScreenBuffer()
	}

	config.PostEventFunc(NewEventFileChanged(length, fm.percentage()))
}

func (fm *FilterManager) ScrollDown() {
	fm.commandChannel <- CommandDown{}
}

func (fm *FilterManager) ScrollUp() {
	fm.commandChannel <- CommandUp{}
}

func (fm *FilterManager) ScrollHorizontal(offset int) {
	fm.commandChannel <- CommandScrollHorizontal{offset}
}

func (fm *FilterManager) PageDown() {
	fm.commandChannel <- CommandPgDown{}
}

func (fm *FilterManager) PageUp() {
	fm.commandChannel <- CommandPgUp{}
}

func (fm *FilterManager) ScrollEnd() {
	fm.commandChannel <- CommandEnd{}
}

func (fm *FilterManager) ScrollHome() {
	fm.commandChannel <- CommandHome{}
}

func (fm *FilterManager) FindMatch(direction filter.ScrollDirection) {
	fm.commandChannel <- CommandFindMatch{direction}
}

func (fm *FilterManager) AddFilter(filter filter.Filter) {
	fm.commandChannel <- CommandAddFilter{filter}
}

func (fm *FilterManager) RemoveFilter(filter filter.Filter) {
	fm.commandChannel <- CommandRemoveFilter{filter}
}

func (fm *FilterManager) SetDisplayHeight(height int) {
	fm.commandChannel <- CommandSetDisplayHeight{height}
}

func (fm *FilterManager) SetCurrentLine(line int) {
	fm.commandChannel <- CommandSetCurrentLine{line}
}

func (fm *FilterManager) UpdateFilterColorIndex(filter filter.Filter, colorIndex uint8) {
	fm.commandChannel <- CommandFilterColorIndexUpdate{filter, colorIndex}
}

func (fm *FilterManager) UpdateFilterMode(filter filter.Filter, mode filter.FilterMode) {
	fm.commandChannel <- CommandFilterModeUpdate{filter, mode}
}

func (fm *FilterManager) UpdateFilterCaseSensitiveUpdate(filter filter.Filter, caseSensitive bool) {
	fm.commandChannel <- CommandFilterCaseSensitiveUpdate{filter, caseSensitive}
}

func (fm *FilterManager) UpdateFilterKey(filter filter.Filter, name string, key string) {
	fm.commandChannel <- CommandFilterKeyUpdate{filter, name, key}
}

func (fm *FilterManager) ToggleFollowMode() {
	fm.commandChannel <- CommandToggleFollowMode{}
}

func (fm *FilterManager) processCommand(command Command) {

	// TODO let all these methods return an error, then send a beep indication
	// through the channel in case of an error

	var err error
	switch command := command.(type) {
	case CommandDown:
		err = fm.internalScrollVertical(filter.DirectionDown)
		config.PostEventFunc(NewEventDisplay(*fm.display))
	case CommandUp:
		err = fm.internalScrollVertical(filter.DirectionUp)
		config.PostEventFunc(NewEventDisplay(*fm.display))
	case CommandScrollHorizontal:
		err = fm.internalScrollHorizontal(command.offset)
	case CommandPgDown:
		err = fm.internalScrollPage(filter.DirectionDown)
		config.PostEventFunc(NewEventDisplay(*fm.display))
	case CommandPgUp:
		err = fm.internalScrollPage(filter.DirectionUp)
		config.PostEventFunc(NewEventDisplay(*fm.display))
	case CommandEnd:
		fm.internalScrollEnd()
		config.PostEventFunc(NewEventDisplay(*fm.display))
	case CommandHome:
		fm.internalScrollHome()
		fm.syncRefreshScreenBuffer()
	case CommandFindMatch:
		if refresh, _ := fm.internalFindNextMatch(command.direction); refresh {
			fm.syncRefreshScreenBuffer()
		} else {
			config.PostEventFunc(NewEventDisplay(*fm.display))
		}
	case CommandAddFilter:
		fm.filters.Add(command.Filter)
		fm.syncRefreshScreenBuffer()
	case CommandRemoveFilter:
		err = fm.filters.Remove(command.Filter)
		fm.syncRefreshScreenBuffer()
	case CommandSetDisplayHeight:
		fm.display.SetHeight(command.Lines)
		config.PostEventFunc(NewEventDisplay(*fm.display))
	case CommandSetCurrentLine:
		fm.internalSetCurrentLine(command.Line)
		fm.syncRefreshScreenBuffer()
	case CommandFilterColorIndexUpdate:
		command.Filter.SetColorIndex(command.ColorIndex)
		fm.filters.InvalidateCaches()
		fm.display.UnsetCurrentMatch()
		fm.asyncRefreshScreenBuffer()
	case CommandFilterModeUpdate:
		stringFilter := command.Filter.(*filter.StringFilter)
		stringFilter.SetMode(command.Mode)
		fm.filters.InvalidateCaches()
		fm.display.UnsetCurrentMatch()
		fm.asyncRefreshScreenBuffer()
	case CommandFilterCaseSensitiveUpdate:
		stringFilter := command.Filter.(*filter.StringFilter)
		err = stringFilter.SetCaseSensitive(command.CaseSensitive)
		fm.filters.InvalidateCaches()
		fm.display.UnsetCurrentMatch()
		fm.asyncRefreshScreenBuffer()
	case CommandFilterKeyUpdate:
		fm.filters.InvalidateCaches()
		err = command.Filter.SetKey(command.Name, command.Key)
		fm.display.UnsetCurrentMatch()
		fm.asyncRefreshScreenBuffer()
	case CommandToggleFollowMode:
		fm.internalToggleFollowMode()
		config.PostEventFunc(NewEventDisplay(*fm.display))
	default:
		log.Panicf("Command %s not implemented!", command.commandString())
	}

	if err == util.ErrOutOfBounds || err == util.ErrNotFound ||
		err == filter.ErrNotEnoughPanels || err == filter.ErrRegex {

		config.PostEventFunc(NewEventError(true, ""))
	} else if err != nil {
		// TODO switch back on, problem was the regex errors ended up here
		// log.Panicf("Unknwon error %v+", err)
	}
}

// ----------------------------------

func (fm *FilterManager) internalScrollVertical(
	direction filter.ScrollDirection) error {

	if fm.display.Height() <= 0 {
		return util.ErrOutOfBounds
	}

	var startLine *lines.Line
	if direction == filter.DirectionDown {
		startLine = fm.display.lastLine()
	} else {
		startLine = fm.display.firstLine()
	}

	if startLine.Status == lines.LineDoesNotExist {
		return util.ErrOutOfBounds
	}

	nextLine, err := fm.filters.FindNonHiddenLine(startLine.No, direction)
	if err != nil {
		return err
	}

	if direction == filter.DirectionDown {
		fm.display.addLineAtBottomRemoveLineAtTop(nextLine)
	} else {
		fm.display.addLineAtTopRemoveLineAtBottom(nextLine)
	}

	fm.currentLine = fm.display.firstLine().No

	return nil
}

func (fm *FilterManager) internalScrollHorizontal(offset int) error {
	width, _ := fm.filters.Size()

	newCol, err := util.InBetween(fm.display.CurrentCol+offset, 0, width)
	if err != nil {
		return util.ErrOutOfBounds
	}

	fm.display.SetCurrentCol(newCol)
	return nil
}

func (fm *FilterManager) alreadyAtTheEnd() bool {
	if fm.filters.Source().IsEmpty() {
		return true
	}

	lastLineOnScreen := fm.display.lastLine()

	if lastLineOnScreen.Status == lines.LineDoesNotExist {
		return true
	}

	_, err := fm.filters.FindNonHiddenLine(lastLineOnScreen.No, 1)
	return err != nil
}

func (fm *FilterManager) internalScrollPage(direction filter.ScrollDirection) error {
	for i := 0; i < fm.display.Height()-1; i++ {
		err := fm.internalScrollVertical(direction)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fm *FilterManager) internalTail() {
	fm.syncRefreshScreenBuffer()
	for {
		err := fm.internalScrollVertical(filter.DirectionDown)
		if err != nil {
			break
		}
	}

	fm.display.Percentage = GetFilterManager().percentage()

	config.PostEventFunc(NewEventDisplay(*fm.display))
}

func (fm *FilterManager) internalScrollEnd() {
	y := fm.display.Height() - 1
	lineNo := fm.filters.SourceLength() - 1
	for ; y >= 0 && lineNo >= 0; lineNo-- {
		line, _ := fm.filters.GetLine(lineNo)
		if line.Status != lines.LineHidden &&
			line.Status != lines.LineDoesNotExist {

			fm.display.Buffer[y] = line
			y--
		}
	}

	if y >= 0 {
		y++
		y0 := 0
		for ; y < fm.display.Height(); y, y0 = y+1, y0+1 {
			fm.display.Buffer[y0] = fm.display.Buffer[y]
		}

		fm.display.fillRestOfBufferWithNonExistingLines(y0)
	}

	// really, does this belong here?
	// and shouldn't we send an event instead?!
	fm.display.Percentage = GetFilterManager().percentage()
}

func (fm *FilterManager) internalScrollHome() {
	fm.internalSetCurrentLine(0)
}

func (fm *FilterManager) syncRefreshScreenBuffer() {
	fm.display.refreshDisplay(context.Background(), nil, fm.currentLine)
}

func (fm *FilterManager) asyncRefreshScreenBuffer() {
	var ctx context.Context

	if fm.refresherCancelFunc != nil {
		fm.refresherCancelFunc()
		fm.refresherCancelFunc = nil
		fm.refresherWg.Wait()
	}

	ctx, fm.refresherCancelFunc = context.WithCancel(fm.ctx)

	fm.refresherWg.Add(1)
	go fm.display.refreshDisplay(ctx, &fm.refresherWg, fm.currentLine)
}

func (fm *FilterManager) internalFindNextMatch(direction filter.ScrollDirection) (bool, error) {
	fail.If(direction != 1 && direction != -1, "Unknown direction %d", direction)

	startSearchWith := 0
	var found *lines.Line

	// first see if the current match is on screen if yes, try if we can find
	// the next match on screen already (much faster)
	screenLine, err := fm.display.getLineOnScreen(fm.display.CurrentMatch)
	if err == nil {
		startSearchWith = fm.display.Buffer[screenLine].No
		found, err = fm.display.searchOnScreen(screenLine+int(direction), direction)
		if err == nil {
			fm.display.CurrentMatch = found.No
			return false, nil
		} else if err != util.ErrNotFound {
			log.Panicf("Unkown error %v+", err)
		}
	} else if len(fm.display.Buffer) > 0 {
		// no match on screen found, start searching through the filters either
		// beginning with the first line on screen or (if nothing's displayed yet -
		// how is this happening?!) start with the beginning of the file
		startSearchWith = fm.display.firstLine().No
	}
	startSearchWith, _ = util.InBetween(startSearchWith+int(direction), 0,
		fm.filters.SourceLength()-1)

	found, err = fm.filters.Search(startSearchWith, direction)
	if err != nil {
		// necessary?
		// fm.display.UnsetCurrentMatch()
		return false, err
	}

	fm.display.CurrentMatch = found.No

	// I think this if statement is not necessary anymore?!
	// if !fm.isLineOnScreen(found.No) {
	var percentage int
	if direction == filter.DirectionDown {
		percentage = 25
	} else {
		percentage = 75
	}
	firstLine, _ := fm.arrangeLine(found.No, percentage)
	fm.internalSetCurrentLine(firstLine)
	// }

	return true, nil
}

func (fm *FilterManager) internalSetCurrentLine(newCurrentLine int) {
	fm.currentLine = newCurrentLine
}

//   - if currently following and display is positioned at the end
//     --> stop following
//   - if currently following but display is somewhere else than at the end
//     --> bring display to the end, continue following
//   - if currently not following
//     --> bring display to the end and start following
func (fm *FilterManager) internalToggleFollowMode() {
	cfg := config.UserCfg()

	if cfg.Follow {
		if fm.alreadyAtTheEnd() {
			if !cfg.Stdin {
				fm.readerCancelFunc()
			}
			cfg.Follow = false
		} else {
			fm.internalScrollEnd()
		}
	} else {
		cfg.Follow = true
		fm.internalTail()
		if !cfg.Stdin {
			fm.wg.Add(1)
			var ctx context.Context
			ctx, fm.readerCancelFunc = context.WithCancel(fm.ctx)
			go reader.GetReader().ReopenForWatching(ctx, fm.wg, cfg.FilePath,
				fm.contentUpdate, fm.filters.Source().LastLine().No+1)
		}
	}
}

func (fm *FilterManager) percentage() int {
	length := fm.filters.SourceLength()
	if length <= 0 || fm.currentLine < 0 || fm.currentLine > length {
		return 0
	}

	percentage := 100 * (fm.currentLine + fm.display.Height()) / length
	if percentage > 100 {
		percentage = 100
	}

	return percentage
}

func (fm *FilterManager) arrangeLine(lineNo int, percentage int) (int, error) {
	fail.If(len(fm.display.Buffer) <= 0, "arrangeLine() called with lineNo=%d but empty buffer?!?", lineNo)

	linesAbove := percentage*len(fm.display.Buffer)/100 - 1

	var line *lines.Line
	var err error
	for i := 1; i <= linesAbove; i++ {
		busy.SpinWithFraction(lineNo, fm.filters.SourceLength())
		line, err = fm.filters.FindNonHiddenLine(lineNo, -1)
		if err != nil {
			return 0, err
		}
	}

	return line.No, nil
}
