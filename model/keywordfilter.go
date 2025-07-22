package model

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type KeywordFilter struct {
	source       Filter
	keyword      string
	eventHandler tcell.EventHandler
}

func NewKeywordFilter(source Filter) *KeywordFilter {
	k := &KeywordFilter{}
	k.source = source

	return k
}

func (k *KeywordFilter) UpdateText(text string) {
	log.Printf("KeywordFilter.UpdateText: %s", text)
	k.keyword = text

	k.HandleEvent(NewEventFilterOutput())
	log.Println(".")
}

func (k *KeywordFilter) GetLine(line int) (Line, error) {
	sourceLine, err := k.source.GetLine(line)

	if err != nil {
		log.Println("Keywordfilter GetLine error")
		return sourceLine, err
	}

	if !strings.Contains(sourceLine.Str, k.keyword) {
		return Line{}, ErrLineDidNotMatch
	}

	return sourceLine, nil
}

func (k *KeywordFilter) Source() (Filter, error) {
	if k.source == nil {
		return nil, fmt.Errorf("no source defined")
	}

	return k.source, nil
}

func (k *KeywordFilter) SetSource(source Filter) {
	k.source = source
}

func (k *KeywordFilter) Size() (int, int, error) {
	//return 80, 0, nil // FIXME
	return k.source.Size()
}

func (k *KeywordFilter) SetEventHandler(eventHandler tcell.EventHandler) {
	log.Printf("Keywordfilter.EventHandler")
	k.eventHandler = eventHandler
}

func (k *KeywordFilter) HandleEvent(ev tcell.Event) bool {
	log.Printf("Firing handleEvents: %T", ev)
	if k.eventHandler == nil {
		log.Printf("but eventHandler not set")
		return false
	}
	return k.eventHandler.HandleEvent(ev)
}
