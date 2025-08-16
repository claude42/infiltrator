package filter

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
