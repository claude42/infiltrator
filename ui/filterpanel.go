package ui

import (
	"github.com/claude42/infiltrator/components"
	"github.com/claude42/infiltrator/config"
)

type FilterPanel interface {
	components.Panel

	SetPanelConfig(panelConfig *config.PanelTable)
}

// FilterPanelImpl is a base struct for filter panels. It must not be used
// directly. Instead embed it into another struct. And make sure these structs
// provide their own implementation of SetPanelConfig() and call this from
// their constructor.
type FilterPanelImpl struct {
	*ColoredPanel

	panelType   config.FilterType
	panelConfig config.PanelTable
}

func NewFilterPanelImpl(panelType config.FilterType, name string) *FilterPanelImpl {
	g := &FilterPanelImpl{
		ColoredPanel: NewColoredPanel(name),
		panelType:    panelType,
	}

	return g
}
