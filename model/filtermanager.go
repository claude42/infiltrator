package model

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model/busy"
	"github.com/claude42/infiltrator/model/filter"
	"github.com/claude42/infiltrator/model/reader"
	"github.com/claude42/infiltrator/util"
)

var (
	filterManagerInstance *FilterManager
	ErrNotEnoughPanels    = errors.New("at least one buffer and one filter required")
)

type FilterManager struct {
	util.ObservableImpl

	ctx  context.Context
	wg   *sync.WaitGroup
	quit chan<- string

	readerCancelFunc context.CancelFunc

	refresherCancelFunc context.CancelFunc
	refresherWg         sync.WaitGroup

	contentUpdate  chan []*reader.Line
	commandChannel chan Command

	filters     []filter.Filter
	currentLine int

	display *Display
}

func GetFilterManager() *FilterManager {
	fail.IfNil(filterManagerInstance, "Filtermanager missing!")

	return filterManagerInstance
}

func NewFilterManager(ctx context.Context, wg *sync.WaitGroup, quit chan<- string) *FilterManager {
	fm := &FilterManager{}
	fm.ctx = ctx
	fm.wg = wg
	fm.quit = quit

	fm.display = &Display{}
	fm.display.UnsetCurrentMatch()

	fm.contentUpdate = make(chan []*reader.Line, 10)
	fm.commandChannel = make(chan Command, 10)

	fm.internalAddFilter(filter.NewSource())
	fm.internalAddFilter(filter.NewDateFilter())
	fm.internalAddFilter(filter.NewCache())

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
		fm.contentUpdate, config.GetConfiguration().FollowFile)
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

func (fm *FilterManager) Source() *filter.Source {
	fm.Lock()
	fail.If(len(fm.filters) < 1, "No source in filter stack!")
	firstFilter := fm.filters[0]
	fm.Unlock()

	source, ok := firstFilter.(*filter.Source)
	fail.If(!ok, "First filter is not a Source!")
	return source
}

func (fm *FilterManager) outputFilter() (filter.Filter, error) {
	fm.Lock()
	defer fm.Unlock()
	if len(fm.filters) == 0 {
		return nil, fmt.Errorf("pipeline empty")
	}

	return fm.filters[len(fm.filters)-1], nil
}

func (fm *FilterManager) size() (int, int) {
	filter, err := fm.outputFilter()
	if err != nil {
		return 0, 0
	}

	return filter.Size()
}

func (fm *FilterManager) sourceLength() int {
	filter, err := fm.outputFilter()
	if err != nil {
		return 0
	}

	return filter.Length()
}

// TODO: make private
func (fm *FilterManager) GetLine(line int) (*reader.Line, error) {
	filter, err := fm.outputFilter()
	if err != nil {
		return reader.NonExistingLine, err
	}

	busy.SpinWithFraction(line, fm.sourceLength())

	return filter.GetLine(line)
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
			// log.Printf("Received contentupdate")
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

func (fm *FilterManager) processContentUpdate(newLines []*reader.Line) {

	// If we're in Follow mode we'll automatically jump to the new end of the
	// file - but only in case we're already at the end
	goToEnd := false
	if config.GetConfiguration().FollowFile && fm.alreadyAtTheEnd() {
		goToEnd = true
	}

	length := fm.Source().StoreNewLines(newLines)
	fm.display.SetTotalLength(length)

	// refresh display as necessary
	if goToEnd {
		// two times refreshDisplay() currently necessary because
		// internalScrollEnd() depends on a correctly set up display.
		// UPD: I think this should not be necessary anymore, let's still keep
		// the comment here
		// fm.refreshDisplay()
		fm.internalScrollEnd()
		fm.syncRefreshScreenBuffer()
	} else if fm.isDisplayAffected() {
		fm.syncRefreshScreenBuffer()
	}

	config.GetConfiguration().PostEventFunc(NewEventFileChanged(length, fm.percentage()))
}

func (fm *FilterManager) isDisplayAffected() bool {
	// only if display is currently at the end and not all lines of the
	// display are filled, then new lines in the file will affect the display.
	if len(fm.display.Buffer) == 0 {
		return true
	}
	return fm.display.Buffer[len(fm.display.Buffer)-1].No == -1
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

func (fm *FilterManager) FindMatch(direction int) {
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
		err = fm.internalScrollDownLineBuffer()
		config.GetConfiguration().PostEventFunc(NewEventDisplay(*fm.display))
	case CommandUp:
		err = fm.internalScrollUpLineBuffer()
		config.GetConfiguration().PostEventFunc(NewEventDisplay(*fm.display))
	case CommandScrollHorizontal:
		err = fm.internalScrollHorizontal(command.offset)
	case CommandPgDown:
		err = fm.internalPageDownLineBuffer()
		config.GetConfiguration().PostEventFunc(NewEventDisplay(*fm.display))
	case CommandPgUp:
		err = fm.internalPageUpLineBuffer()
		config.GetConfiguration().PostEventFunc(NewEventDisplay(*fm.display))
	case CommandEnd:
		fm.internalScrollEnd()
		config.GetConfiguration().PostEventFunc(NewEventDisplay(*fm.display))
	case CommandHome:
		fm.internalScrollHome()
		fm.syncRefreshScreenBuffer()
	case CommandFindMatch:
		if refresh, _ := fm.internalFindNextMatch(command.direction); refresh {
			fm.syncRefreshScreenBuffer()
		} else {
			config.GetConfiguration().PostEventFunc(NewEventDisplay(*fm.display))
		}
	case CommandAddFilter:
		fm.internalAddFilter(command.Filter)
		fm.syncRefreshScreenBuffer()
	case CommandRemoveFilter:
		err = fm.internalRemoveFilter(command.Filter)
		fm.syncRefreshScreenBuffer()
	case CommandSetDisplayHeight:
		fm.display.SetHeight(command.Lines)
		config.GetConfiguration().PostEventFunc(NewEventDisplay(*fm.display))
	case CommandSetCurrentLine:
		fm.internalSetCurrentLine(command.Line)
		fm.syncRefreshScreenBuffer()
	case CommandFilterColorIndexUpdate:
		command.Filter.SetColorIndex(command.ColorIndex)
		fm.invalidateCaches()
		fm.display.UnsetCurrentMatch()
		fm.asyncRefreshScreenBuffer()
	case CommandFilterModeUpdate:
		command.Filter.SetMode(command.Mode)
		fm.invalidateCaches()
		fm.display.UnsetCurrentMatch()
		fm.asyncRefreshScreenBuffer()
	case CommandFilterCaseSensitiveUpdate:
		err = command.Filter.SetCaseSensitive(command.CaseSensitive)
		fm.invalidateCaches()
		fm.display.UnsetCurrentMatch()
		fm.asyncRefreshScreenBuffer()
	case CommandFilterKeyUpdate:
		fm.invalidateCaches()
		err = command.Filter.SetKey(command.Name, command.Key)
		fm.display.UnsetCurrentMatch()
		fm.asyncRefreshScreenBuffer()
	case CommandToggleFollowMode:
		fm.internalToggleFollowMode()
		config.GetConfiguration().PostEventFunc(NewEventDisplay(*fm.display))
	default:
		log.Panicf("Command %s not implemented!", command.commandString())
	}

	if err == util.ErrOutOfBounds || err == util.ErrNotFound ||
		err == ErrNotEnoughPanels || err == filter.ErrRegex {

		config.GetConfiguration().PostEventFunc(NewEventError(true, ""))
	} else if err != nil {
		// TODO switch back on, problem was the regex errors ended up here
		// log.Panicf("Unknwon error %v+", err)
	}
}

// ----------------------------------

func (fm *FilterManager) internalScrollDownLineBuffer() error {
	var nextLine *reader.Line
	var err error

	if fm.display.Height() <= 0 {
		return util.ErrOutOfBounds
	}

	lastLineOnScreen := fm.display.Buffer[fm.display.Height()-1]
	if lastLineOnScreen.Status == reader.LineDoesNotExist {
		return util.ErrOutOfBounds
	}

	lineNo := lastLineOnScreen.No + 1

	for ; ; lineNo++ {
		nextLine, err = fm.GetLine(lineNo)
		if err != nil {
			return util.ErrOutOfBounds
		} else if nextLine.Status == reader.LineWithoutStatus ||
			nextLine.Status == reader.LineMatched ||
			nextLine.Status == reader.LineDimmed {

			break
		}
	}

	if fm.display.Height() > 0 {
		fm.display.Buffer = append(fm.display.Buffer[1:], nextLine)
	} else {
		fm.display.Buffer = []*reader.Line{nextLine}
	}

	fm.currentLine = fm.display.Buffer[0].No

	return nil
}

func (fm *FilterManager) internalScrollHorizontal(offset int) error {
	width, _ := fm.size()

	newCol, err := util.InBetween(fm.display.CurrentCol+offset, 0, width)
	if err != nil {
		return util.ErrOutOfBounds
	}

	fm.display.SetCurrentCol(newCol)
	return nil
}

func (fm *FilterManager) alreadyAtTheEnd() bool {
	if fm.Source().IsEmpty() {
		return true
	}

	lastLineOnScreen := fm.display.Buffer[len(fm.display.Buffer)-1]

	if lastLineOnScreen.Status == reader.LineDoesNotExist {
		return true
	}

	lineNo := lastLineOnScreen.No + 1

	// code duplication with internalScrollDownLineBuffer
	for ; ; lineNo++ {
		nextLine, err := fm.GetLine(lineNo)
		if err != nil {
			return true
		}
		if nextLine.Status == reader.LineWithoutStatus ||
			nextLine.Status == reader.LineMatched ||
			nextLine.Status == reader.LineDimmed {

			return false
		}
	}
}

// will return the line it scrolled to
func (fm *FilterManager) internalScrollUpLineBuffer() error {
	var prevLine *reader.Line
	var err error

	lineNo := fm.display.Buffer[0].No - 1

	if lineNo < 0 {
		return util.ErrOutOfBounds
	}

	for ; lineNo >= 0; lineNo-- {
		prevLine, err = fm.GetLine(lineNo)
		if err != nil || prevLine.Status == reader.LineWithoutStatus ||
			prevLine.Status == reader.LineMatched ||
			prevLine.Status == reader.LineDimmed {

			// matching line found
			break
		}
	}

	if lineNo < 0 {
		return util.ErrOutOfBounds
	}

	if fm.display.Height() > 0 {
		fm.display.Buffer = append([]*reader.Line{prevLine},
			fm.display.Buffer[:fm.display.Height()-1]...)
	} else {
		fm.display.Buffer = []*reader.Line{prevLine}
	}

	fm.currentLine = fm.display.Buffer[0].No

	return nil
}

func (fm *FilterManager) internalPageDownLineBuffer() error {
	for i := 0; i < fm.display.Height()-1; i++ {
		err := fm.internalScrollDownLineBuffer()
		if err != nil {
			return err
		}
	}
	return nil
}

func (fm *FilterManager) internalPageUpLineBuffer() error {
	for i := 0; i < fm.display.Height()-1; i++ {
		err := fm.internalScrollUpLineBuffer()
		if err != nil {
			return err
		}
	}
	return nil
}

func (fm *FilterManager) internalScrollEnd() {
	y := fm.display.Height() - 1
	lineNo := fm.sourceLength() - 1
	for ; y >= 0 && lineNo >= 0; lineNo-- {
		line, _ := fm.GetLine(lineNo)
		if line.Status != reader.LineHidden &&
			line.Status != reader.LineDoesNotExist {

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

func (fm *FilterManager) internalAddFilter(f filter.Filter) {
	fm.Lock()
	defer fm.Unlock()

	if len(fm.filters) == 0 {
		fm.filters = append(fm.filters, f)
		return
	}

	var pos int
	for pos = len(fm.filters) - 1; pos > 0; pos-- {
		existingFilter := fm.filters[pos]
		if _, ok := existingFilter.(*filter.Cache); ok {
			continue
		} else if _, ok := existingFilter.(*filter.Source); ok {
			log.Panic("really?!?")
		} else {
			break
		}
	}

	// if the whole for loop went through then pos should be 0 now. So new
	// filter will be added add pos 1, right after the source.

	f.SetSource(fm.filters[pos])
	if pos < len(fm.filters)-1 {
		fm.filters[pos+1].SetSource(f)
		f.SetSource(fm.filters[pos])
		// insert right after fm.filters[pos]
		fm.filters = append(fm.filters[:pos+2], fm.filters[pos+1:]...)
		fm.filters[pos+1] = f
	} else {
		fm.filters = append(fm.filters, f)
	}
}

func (fm *FilterManager) internalRemoveFilter(f filter.Filter) error {
	fm.Lock()
	defer fm.Unlock()
	if len(fm.filters) <= 2 {
		return ErrNotEnoughPanels
	}
	for i, filter := range fm.filters {
		if filter == f {
			fail.If(i == 0, "Cannot remove source from pipeline")
			fm.filters = append(fm.filters[:i], fm.filters[i+1:]...)
			if i < len(fm.filters) {
				fm.filters[i].SetSource(fm.filters[i-1])
			}
			return nil
		}
	}
	log.Panicln("Filter not found in pipeline")
	return fmt.Errorf("Filter not found in pipeline")
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

func (fm *FilterManager) internalFindNextMatch(direction int) (bool, error) {
	fail.If(direction != 1 && direction != -1, "Unknown direction %d", direction)

	startSearchWith := 0
	var found *reader.Line

	// first see if the current match is on screen if yes, try if we can find
	// the next match on screen already (much faster)
	screenLine, err := fm.getLineOnScreen(fm.display.CurrentMatch)
	if err == nil {
		startSearchWith = fm.display.Buffer[screenLine].No
		found, err = fm.searchOnScreen(screenLine+direction, direction)
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
		startSearchWith = fm.display.Buffer[0].No
	}
	startSearchWith, _ = util.InBetween(startSearchWith+direction, 0, fm.sourceLength()-1)

	found, err = fm.search(startSearchWith, direction)
	if err != nil {
		// necessary?
		// fm.display.UnsetCurrentMatch()
		return false, err
	}

	fm.display.CurrentMatch = found.No

	// I think this if statement is not necessary anymore?!
	// if !fm.isLineOnScreen(found.No) {
	var percentage int
	if direction == 1 {
		percentage = 25
	} else {
		percentage = 75
	}
	firstLine, _ := fm.arrangeLine(found.No, percentage)
	fm.internalSetCurrentLine(firstLine)
	// }

	return true, nil
}

func (fm *FilterManager) getLineOnScreen(lineNo int) (int, error) {
	if lineNo < 0 {
		return -1, util.ErrOutOfBounds
	}

	for screenLine, line := range fm.display.Buffer {
		if lineNo == line.No {
			return screenLine, nil
		}
	}
	return -1, util.ErrOutOfBounds
}

func (fm *FilterManager) search(start int, direction int) (*reader.Line, error) {
	length := fm.sourceLength()

	for i := start; ; i = i + direction {
		busy.SpinWithFraction(i, fm.sourceLength())
		newLine, err := fm.GetLine(i)
		if err != nil || i < 0 || i >= length {
			return nil, util.ErrNotFound
		} else if newLine.Matched {
			return newLine, nil
		}
	}
}

func (fm *FilterManager) searchOnScreen(startOnScreen int, direction int) (*reader.Line, error) {
	height := len(fm.display.Buffer)

	for i := startOnScreen; i >= 0 && i < height; i = i + direction {
		if fm.display.Buffer[i].Matched {
			return fm.display.Buffer[i], nil
		}
	}

	return nil, util.ErrNotFound
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
	cfg := config.GetConfiguration()
	// TODO: handle stdin case
	if cfg.FollowFile {
		if fm.alreadyAtTheEnd() {
			fm.readerCancelFunc()
			cfg.FollowFile = false
		} else {
			fm.internalScrollEnd()
		}
	} else {
		cfg.FollowFile = true
		fm.wg.Add(1)
		fm.internalScrollEnd()
		var ctx context.Context
		ctx, fm.readerCancelFunc = context.WithCancel(fm.ctx)
		go reader.GetReader().ReopenForWatching(ctx, fm.wg, cfg.FilePath,
			fm.contentUpdate, fm.Source().LastLine().No+1)
	}
}

func (fm *FilterManager) percentage() int {
	length := fm.sourceLength()
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

	var err error
	for i := 1; i <= linesAbove; i++ {
		busy.SpinWithFraction(lineNo, fm.sourceLength())
		lineNo, err = fm.findNonHiddenLine(lineNo, -1)
		if err != nil {
			return 0, err
		}
	}

	return lineNo, nil
}

func (fm *FilterManager) findNonHiddenLine(lineNo int, direction int) (int, error) {
	fail.If(direction != -1 && direction != 1, "Unknown directionn %d", direction)

	length := fm.sourceLength()

	for lineNo = lineNo + direction; lineNo >= 0 && lineNo < length; lineNo = lineNo + direction {
		busy.SpinWithFraction(lineNo, fm.sourceLength())
		prevLine, err := fm.GetLine(lineNo)
		if err != nil {
			return -1, err
		}
		if prevLine.Status == reader.LineWithoutStatus ||
			prevLine.Status == reader.LineMatched ||
			prevLine.Status == reader.LineDimmed {
			return prevLine.No, nil
		}
	}

	return -1, util.ErrOutOfBounds
}

func (fm *FilterManager) invalidateCaches() {
	for _, f := range fm.filters {
		cache, ok := f.(*filter.Cache)
		if ok {
			cache.Invalidate()
		}
	}
}

func (fm *FilterManager) GetDateFilter() (*filter.DateFilter, error) {
	for _, f := range fm.filters {
		dateFilter, ok := f.(*filter.DateFilter)
		if ok {
			return dateFilter, nil
		}
	}

	return nil, util.ErrNotFound
}
