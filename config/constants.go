package config

import "github.com/claude42/infiltrator/util"

type FilterType int

const (
	appName            = "infiltrator"
	mainConfigFileName = "/config.toml"
	historyFileName    = "/history.toml"
	presetDir          = "/presets/"
)

const (
	FilterTypeKeyword FilterType = iota
	FilterTypeRegex
	FilterTypeDate
	PanelTypeGlob
	PanelTypeHost
	PanelTypeFacility

	FilterTypeCount

	FilterStringKeyword = "Keyword"
	FilterStringRegex   = "Regex"
	FilterStringDate    = "Date"

	// These are no filters of their own, just names for the two inputs for
	// the date filter
	FilterStringFrom = "From"
	FilterStringTo   = "To"
)

type FilterTuple struct {
	FilterType   FilterType
	FilterString string
}

type FilterSlice []FilterTuple

var Filters FilterSlice = []FilterTuple{
	{FilterType: FilterTypeKeyword, FilterString: FilterStringKeyword},
	{FilterType: FilterTypeRegex, FilterString: FilterStringRegex},
	{FilterType: FilterTypeDate, FilterString: FilterStringDate},
}

func (FilterSlice) String(key FilterType) (string, error) {
	for _, v := range Filters {
		if v.FilterType == key {
			return v.FilterString, nil
		}
	}
	return "", util.ErrNotFound
}

func (FilterSlice) AllStrings() []string {
	var result []string
	for _, v := range Filters {
		result = append(result, v.FilterString)
	}
	return result
}

func (FilterSlice) Type(value string) (FilterType, error) {
	for _, v := range Filters {
		if v.FilterString == value {
			return v.FilterType, nil
		}
	}
	return -1, util.ErrNotFound
}

var Histories []string = []string{
	FilterStringKeyword,
	FilterStringRegex,
	FilterStringFrom,
	FilterStringTo,
}

type FilterMode int

var FilterModeStrings = []string{
	"focus",
	"match",
	"hide",
}

var FilterModeNames = []string{
	"focus",
	"match",
	"hide",
}

const (
	FilterFocus FilterMode = iota
	FilterMatch
	FilterHide
)

func (fm FilterMode) String() string {
	return FilterModeStrings[fm]
}

var CaseSensitiveStrings = []string{
	"case",
	"CaSe",
}

// ----------------------------------------------------------------

const PanelNameWidth = 11
const PanelHeaderWidth = 26
const PanelHeaderGap = 2
