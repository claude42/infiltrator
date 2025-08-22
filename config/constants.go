package config

import "github.com/claude42/infiltrator/util"

type FilterType int

const (
	appName            = "infiltrator"
	mainConfigFileName = "/config.toml"
	presetDir          = "/presets/"
)

const (
	FilterTypeKeyword FilterType = iota
	FilterTypeRegex
	PanelTypeGlob
	PanelTypeHost
	PanelTypeFacility
	FilterTypeDate

	FilterStringKeyword = "Keyword"
	FilterStringRegex   = "Regex"
	FilterStringDate    = "Date"

	// These are no filters of their own, just names for the two inputs for
	// the date filter
	FilterStringFrom = "From"
	FilterStringTo   = "To"
)

// TODO: some stuff is here, some in stringfiltermodes.go
var Filters map[FilterType]string = map[FilterType]string{
	FilterTypeKeyword: FilterStringKeyword,
	FilterTypeRegex:   FilterStringRegex,
	FilterTypeDate:    FilterStringDate,
}

var Histories []string = []string{
	FilterStringKeyword,
	FilterStringRegex,
	FilterStringFrom,
	FilterStringTo,
}

func FilterNameToType(filterName string) (FilterType, error) {
	for key, value := range Filters {
		if value == filterName {
			return key, nil
		}
	}
	return -1, util.ErrNotFound
}
