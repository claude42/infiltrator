package config

type FilterType int

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

	// These are no filters of there own, just names for the two inputs for
	// the date filter
	FilterStringFrom = "From"
	FilterStringTo   = "To"
)

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
