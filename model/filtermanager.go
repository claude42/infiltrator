package model

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model/busy"
	"github.com/claude42/infiltrator/util"
)

var (
	filterManagerInstance *FilterManager
	filterManagerOnce     sync.Once
	ErrNotEnoughPanels    = errors.New("at least one buffer and one filter required")
)

type FilterManager struct {
	util.ObservableImpl

	readerContext context.Context
	readerCancel  context.CancelFunc

	refresherCancelFunc context.CancelFunc
	refesherWg          sync.WaitGroup

	contentUpdate  chan []*Line
	commandChannel chan Command

	filters     []Filter
	currentLine int

	display *Display
}

func GetFilterManager() *FilterManager {
	filterManagerOnce.Do(func() {
		filterManagerInstance = createNewFilterManager()
	})
	return filterManagerInstance
}

func createNewFilterManager() *FilterManager {
	fm := &FilterManager{}
	fm.display = &Display{}
	fm.display.UnsetCurrentMatch()
	fm.contentUpdate = make(chan []*Line, 10)
	fm.commandChannel = make(chan Command, 10)

	fm.internalAddFilter(&Source{})
	fm.internalAddFilter(NewCache())
	return fm
}

func (fm *FilterManager) ReadFromFile(filePath string) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("A panic occurred: %v\nStack trace:\n%s", r, debug.Stack())
			panic(r)
		}
	}()

	fm.readerContext, fm.readerCancel = context.WithCancel(config.GetConfiguration().Context)
	config.GetConfiguration().WaitGroup.Add(1)
	go GetReader().ReadFromFile(filePath, fm.readerContext, fm.contentUpdate, config.GetConfiguration().FollowFile)
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
	go GetReader().ReadFromStdin(fm.contentUpdate)
	// GetLoremIpsumReader().Read(fm.contentUpdate)
}

func (fm *FilterManager) Source() *Source {
	fm.Lock()
	if len(fm.filters) < 1 {
		log.Panic("No source in filter stack!")
	}
	firstFilter := fm.filters[0]
	fm.Unlock()

	source, ok := firstFilter.(*Source)
	if !ok {
		log.Panic("First filter is not a Source!")
	}
	return source
}

func (fm *FilterManager) outputFilter() (Filter, error) {
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

	return filter.size()
}

func (fm *FilterManager) sourceLength() int {
	filter, err := fm.outputFilter()
	if err != nil {
		return 0
	}

	return filter.length()
}

// TODO: make private
func (fm *FilterManager) GetLine(line int) (*Line, error) {
	filter, err := fm.outputFilter()
	if err != nil {
		return &Line{}, err
	}

	busy.SpinWithFraction(line, fm.sourceLength())

	return filter.getLine(line)
}

func (fm *FilterManager) EventLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("A panic occurred: %v\nStack trace:\n%s", r, debug.Stack())
			panic(r)
		}
	}()

	cfg := config.GetConfiguration()
	defer cfg.WaitGroup.Done()

	for {
		select {
		case newLines := <-fm.contentUpdate:
			// log.Printf("Received contentupdate")
			fm.processContentUpdate(newLines)
		case command := <-fm.commandChannel:
			log.Printf("Received Command: %T", command)
			fm.processCommand(command)
		case <-cfg.Context.Done():
			log.Println("Received shutdown")
			return
		}
	}
}

func (fm *FilterManager) processContentUpdate(newLines []*Line) {

	// If we're in Follow mode we'll automatically jump to the new end of the
	// file - but only in case we're already at the end
	goToEnd := false
	if config.GetConfiguration().FollowFile && fm.alreadyAtTheEnd() {
		goToEnd = true
	}

	length := fm.Source().storeNewLines(newLines)
	fm.display.SetTotalLength(length)

	// refresh display as necessary
	if goToEnd {
		// two times refreshDisplay() currently necessary because
		// internalScrollEnd() depends on a correctly set up display.
		// UPD: I think this should not be necessary anymore, let's still keep
		// the comment here
		// fm.refreshDisplay()
		fm.internalScrollEnd()
		fm.display.refreshDisplay(nil, nil, fm.currentLine, false)
	} else if fm.isDisplayAffected() {
		fm.display.refreshDisplay(nil, nil, fm.currentLine, false)
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

func (fm *FilterManager) ScrollDown() (Line, error) {
	fm.commandChannel <- CommandDown{}
	return Line{}, nil
}

func (fm *FilterManager) ScrollUp() (Line, error) {
	fm.commandChannel <- CommandUp{}
	return Line{}, nil
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

func (fm *FilterManager) AddFilter(filter Filter) {
	fm.commandChannel <- CommandAddFilter{filter}
}

func (fm *FilterManager) RemoveFilter(filter Filter) {
	fm.commandChannel <- CommandRemoveFilter{filter}
}

func (fm *FilterManager) SetDisplayHeight(height int) {
	fm.commandChannel <- CommandSetDisplayHeight{height}
}

func (fm *FilterManager) SetCurrentLine(line int) {
	fm.commandChannel <- CommandSetCurrentLine{line}
}

func (fm *FilterManager) UpdateFilterColorIndex(filter Filter, colorIndex uint8) {
	fm.commandChannel <- CommandFilterColorIndexUpdate{filter, colorIndex}
}

func (fm *FilterManager) UpdateFilterMode(filter Filter, mode FilterMode) {
	fm.commandChannel <- CommandFilterModeUpdate{filter, mode}
}

func (fm *FilterManager) UpdateFilterCaseSensitiveUpdate(filter Filter, caseSensitive bool) {
	fm.commandChannel <- CommandFilterCaseSensitiveUpdate{filter, caseSensitive}
}

func (fm *FilterManager) UpdateFilterKey(filter Filter, key string) {
	fm.commandChannel <- CommandFilterKeyUpdate{filter, key}
}

func (fm *FilterManager) ToggleFollowMode() {
	fm.commandChannel <- CommandToggleFollowMode{}
}

func (fm *FilterManager) processCommand(command Command) {
	refreshScreenBuffer := true
	unsetCurrentMatch := false

	// TODO let all these methods return an error, then send a beep indication
	// through the channel in case of an error

	var err error
	switch command := command.(type) {
	case CommandDown:
		err = fm.internalScrollDownLineBuffer()
	case CommandUp:
		err = fm.internalScrollUpLineBuffer()
	case CommandScrollHorizontal:
		err = fm.internalScrollHorizontal(command.offset)
	case CommandPgDown:
		err = fm.internalPageDownLineBuffer()
	case CommandPgUp:
		err = fm.internalPageUpLineBuffer()
	case CommandEnd:
		fm.internalScrollEnd()
	case CommandHome:
		fm.internalScrollHome()
	case CommandFindMatch:
		err = fm.internalFindNextMatch(command.direction)
	case CommandAddFilter:
		fm.internalAddFilter(command.Filter)
	case CommandRemoveFilter:
		err = fm.internalRemoveFilter(command.Filter)
	case CommandSetDisplayHeight:
		fm.display.SetHeight(command.Lines)
	case CommandSetCurrentLine:
		fm.internalSetCurrentLine(command.Line)
	case CommandFilterColorIndexUpdate:
		command.Filter.setColorIndex(command.ColorIndex)
		fm.invalidateCaches()
		unsetCurrentMatch = true
		refreshScreenBuffer = false
		fm.asyncRefreshScreenBuffer(unsetCurrentMatch)
	case CommandFilterModeUpdate:
		command.Filter.setMode(command.Mode)
		fm.invalidateCaches()
		unsetCurrentMatch = true
		refreshScreenBuffer = false
		fm.asyncRefreshScreenBuffer(unsetCurrentMatch)
	case CommandFilterCaseSensitiveUpdate:
		err = command.Filter.setCaseSensitive(command.CaseSensitive)
		fm.invalidateCaches()
		unsetCurrentMatch = true
		refreshScreenBuffer = false
		fm.asyncRefreshScreenBuffer(unsetCurrentMatch)
	case CommandFilterKeyUpdate:
		log.Println("Invalidating caches")
		fm.invalidateCaches()
		log.Println("Setting key")
		err = command.Filter.setKey(command.Key)
		log.Println("Unsetting Current Match")
		unsetCurrentMatch = true
		refreshScreenBuffer = false
		fm.asyncRefreshScreenBuffer(unsetCurrentMatch)
	case CommandToggleFollowMode:
		fm.internalToggleFollowMode()
	default:
		log.Panicf("Command %s not implemented!", command.commandString())
	}
	// Really for every command?
	if refreshScreenBuffer {
		fm.display.refreshDisplay(nil, nil, fm.currentLine, unsetCurrentMatch)

	}
	if err == util.ErrOutOfBounds || err == util.ErrNotFound ||
		err == ErrNotEnoughPanels || err == ErrRegex {

		config.GetConfiguration().PostEventFunc(NewEventError(true, ""))
	} else if err != nil {
		// TODO switch back on
		// log.Panicf("Unknwon error %v+", err)
	}
}

// ----------------------------------

// will return the line it scrolled to
func (fm *FilterManager) internalScrollDownLineBuffer() error {
	var nextLine *Line
	var err error

	if fm.display.Height() <= 0 {
		return util.ErrOutOfBounds
	}

	lastLineOnScreen := fm.display.Buffer[fm.display.Height()-1]
	if lastLineOnScreen.Status == LineDoesNotExist {
		return util.ErrOutOfBounds
	}

	lineNo := lastLineOnScreen.No + 1

	for ; ; lineNo++ {
		nextLine, err = fm.GetLine(lineNo)
		if err != nil {
			return util.ErrOutOfBounds
		} else if nextLine.Status == LineWithoutStatus ||
			nextLine.Status == LineMatched || nextLine.Status == LineDimmed {

			break
		}
	}

	if fm.display.Height() > 0 {
		fm.display.Buffer = append(fm.display.Buffer[1:], nextLine)
	} else {
		fm.display.Buffer = []*Line{nextLine}
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
	if fm.Source().isEmpty() {
		return true
	}

	lastLineOnScreen := fm.display.Buffer[len(fm.display.Buffer)-1]

	if lastLineOnScreen.Status == LineDoesNotExist {
		return true
	}

	lineNo := lastLineOnScreen.No + 1

	// code duplication with internalScrollDownLineBuffer
	for ; ; lineNo++ {
		nextLine, err := fm.GetLine(lineNo)
		if err != nil {
			return true
		}
		if nextLine.Status == LineWithoutStatus ||
			nextLine.Status == LineMatched || nextLine.Status == LineDimmed {

			return false
		}
	}
}

// will return the line it scrolled to
func (fm *FilterManager) internalScrollUpLineBuffer() error {
	var prevLine *Line
	var err error

	lineNo := fm.display.Buffer[0].No - 1

	if lineNo < 0 {
		return util.ErrOutOfBounds
	}

	for ; lineNo >= 0; lineNo-- {
		prevLine, err = fm.GetLine(lineNo)
		if err != nil || prevLine.Status == LineWithoutStatus ||
			prevLine.Status == LineMatched || prevLine.Status == LineDimmed {

			// matching line found
			break
		}
	}

	// Could have also checked for err, not sure
	// what's more elegant...
	// Kept here just in case the change broke something
	// if lineNo < 0 {
	// 	log.Println("fm.ScrollUp ErrOutOfBounds")
	// 	return util.ErrOutOfBounds
	// }

	if fm.display.Height() > 0 {
		fm.display.Buffer = append([]*Line{prevLine}, fm.display.Buffer[:fm.display.Height()-1]...)
	} else {
		fm.display.Buffer = []*Line{prevLine}
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
	// start with what might be the first line, add 1 to make the following
	// for loop nicer
	firstTry := fm.sourceLength() - len(fm.display.Buffer) + 1

	// this should be enough initialization to make internalScrollUpLinBuffer()
	// work
	fm.currentLine = firstTry
	fm.display.refreshDisplay(nil, nil, fm.currentLine, false)
	// fm.display.Buffer[0] = Line{No: firstTry}

	// scroll until the last line of the screen is non-empty or we're at line 0
	for {
		err := fm.internalScrollUpLineBuffer()
		lastLine := fm.display.Buffer[len(fm.display.Buffer)-1]

		if err != nil || lastLine.Status == LineWithoutStatus ||
			lastLine.Status == LineMatched || lastLine.Status == LineDimmed {

			break
		}
	}
}

func (fm *FilterManager) internalScrollHome() {
	fm.internalSetCurrentLine(0)
}

func (fm *FilterManager) internalAddFilter(f Filter) {
	fm.Lock()
	defer fm.Unlock()

	if len(fm.filters) == 0 {
		fm.filters = append(fm.filters, f)
		return
	}

	var pos int
	for pos = len(fm.filters) - 1; pos > 0; pos-- {
		existingFilter := fm.filters[pos]
		if _, ok := existingFilter.(*Cache); ok {
			continue
		} else if _, ok := existingFilter.(*Source); ok {
			log.Panic("really?!?")
		} else {
			break
		}
	}

	// if the whole for loop went through then pos should be 0 now. So new
	// filter will be added add pos 1, right after the source.

	f.setSource(fm.filters[pos])
	if pos < len(fm.filters)-1 {
		fm.filters[pos+1].setSource(f)
		f.setSource(fm.filters[pos])
		// insert right after fm.filters[pos]
		fm.filters = append(fm.filters[:pos+2], fm.filters[pos+1:]...)
		fm.filters[pos+1] = f
	} else {
		fm.filters = append(fm.filters, f)
	}
}

func (fm *FilterManager) internalRemoveFilter(f Filter) error {
	fm.Lock()
	defer fm.Unlock()
	if len(fm.filters) <= 2 {
		return ErrNotEnoughPanels
	}
	for i, filter := range fm.filters {
		if filter == f {
			if i == 0 {
				log.Panicln("Cannot remove source from pipeline")
			}
			fm.filters = append(fm.filters[:i], fm.filters[i+1:]...)
			if i < len(fm.filters) {
				fm.filters[i].setSource(fm.filters[i-1])
			}
			return nil
		}
	}
	log.Panicln("Filter not found in pipeline")
	return fmt.Errorf("Filter not found in pipeline")
}

func (fm *FilterManager) asyncRefreshScreenBuffer(unsetCurrentMatch bool) {
	var ctx context.Context

	if fm.refresherCancelFunc != nil {
		fm.refresherCancelFunc()
		fm.refresherCancelFunc = nil
		fm.refesherWg.Wait()
	}

	ctx, fm.refresherCancelFunc = context.WithCancel(config.GetConfiguration().Context)

	fm.refesherWg.Add(1)
	go fm.display.refreshDisplay(ctx, &fm.refesherWg, fm.currentLine, unsetCurrentMatch)
}

func (fm *FilterManager) internalFindNextMatch(direction int) error {
	if direction != 1 && direction != -1 {
		log.Panicf("Unknown direction %d", direction)
	}

	startSearchWith := 0
	var found *Line

	// first see if the current match is on screen if yes, try if we can find
	// the next match on screen already (much faster)
	screenLine, err := fm.getLineOnScreen(fm.display.CurrentMatch)
	if err == nil {
		startSearchWith = fm.display.Buffer[screenLine].No
		found, err = fm.searchOnScreen(screenLine+direction, direction)
		if err == nil {
			fm.display.CurrentMatch = found.No
			return nil
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
		return err
	}

	fm.display.CurrentMatch = found.No

	if !fm.isLineOnScreen(found.No) {
		var percentage int
		if direction == 1 {
			percentage = 25
		} else {
			percentage = 75
		}
		firstLine, _ := fm.arrangeLine(found.No, percentage)
		fm.internalSetCurrentLine(firstLine)
	}

	return nil
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

func (fm *FilterManager) isLineOnScreen(lineNo int) bool {
	if lineNo < 0 {
		return false
	}

	for _, line := range fm.display.Buffer {
		if lineNo == line.No {
			return true
		}
	}
	return false
}
func (fm *FilterManager) search(start int, direction int) (*Line, error) {
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

func (fm *FilterManager) searchOnScreen(startOnScreen int, direction int) (*Line, error) {
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
			fm.readerCancel()
			cfg.FollowFile = false
		} else {
			fm.internalScrollEnd()
		}
	} else {
		cfg.FollowFile = true
		cfg.WaitGroup.Add(1)
		fm.internalScrollEnd()
		go GetReader().ReopenForWatching(cfg.FilePath, fm.readerContext,
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
	if len(fm.display.Buffer) <= 0 {
		log.Panicf("arrangeLine() called with lineNo=%d but empty buffer?!?", lineNo)
	}

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
	if direction != -1 && direction != 1 {
		log.Panicf("Unknown direction %d", direction)
	}

	length := fm.sourceLength()

	for lineNo = lineNo + direction; lineNo >= 0 && lineNo < length; lineNo = lineNo + direction {
		busy.SpinWithFraction(lineNo, fm.sourceLength())
		prevLine, err := fm.GetLine(lineNo)
		if err != nil {
			return -1, err
		}
		if prevLine.Status == LineWithoutStatus || prevLine.Status == LineMatched ||
			prevLine.Status == LineDimmed {
			return prevLine.No, nil
		}
	}

	return -1, util.ErrOutOfBounds
}

func (fm *FilterManager) invalidateCaches() {
	for _, filter := range fm.filters {
		cache, ok := filter.(*Cache)
		if ok {
			cache.Invalidate()
		}
	}
}
