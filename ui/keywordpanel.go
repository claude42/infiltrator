package ui

import (
	//"fmt"
	"log"

	"github.com/claude42/infiltrator/model"
	//"github.com/claude42/infiltrator/util"
	//"github.com/gdamore/tcell/v2"
)

type KeywordPanel struct {
	filter model.Filter
	TextEntryPanel
}

func NewKeywordPanel() *KeywordPanel {
	k := &KeywordPanel{}

	filter := &model.KeywordFilter{}
	model.GetPipeline().AddFilter(filter)
	k.filter = filter
	k.TextEntryPanel = *NewTextEntryPanel()
	k.input.SetReceiver(filter)
	log.Println(".")

	return k
}
