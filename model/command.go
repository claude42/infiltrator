package model

type Command interface {
	commandString() string
}

type CommandNone struct {
}

func (d CommandNone) commandString() string {
	return "None"
}

type CommandDown struct {
}

func (d CommandDown) commandString() string {
	return "Commandwn"
}

type CommandUp struct {
}

func (d CommandUp) commandString() string {
	return "Up"
}

type CommandPgDown struct {
}

func (d CommandPgDown) commandString() string {
	return "PgCommandwn"
}

type CommandPgUp struct {
}

func (d CommandPgUp) commandString() string {
	return "PgUp"
}

type CommandEnd struct {
}

func (d CommandEnd) commandString() string {
	return "End"
}

type CommandHome struct {
}

func (d CommandHome) commandString() string {
	return "Home"
}

type CommandFindMatch struct {
	direction int
}

func (d CommandFindMatch) commandString() string {
	return "FindMatch"
}

type CommandAddFilter struct {
	Filter Filter
}

func (d CommandAddFilter) commandString() string {
	return "AddFilter"
}

type CommandRemoveFilter struct {
	Filter Filter
}

func (d CommandRemoveFilter) commandString() string {
	return "RemoveFilter"
}

type CommandSetDisplayHeight struct {
	Lines int
}

func (d CommandSetDisplayHeight) commandString() string {
	return "SetDisplayHeight"
}

type CommandSetCurrentLine struct {
	Line int
}

func (d CommandSetCurrentLine) commandString() string {
	return "SetCurrentLine"
}

type CommandFilterColorIndexUpdate struct {
	Filter     Filter
	ColorIndex uint8
}

func (d CommandFilterColorIndexUpdate) commandString() string {
	return "ColorIndexUpdate"
}

type CommandFilterModeUpdate struct {
	Filter Filter
	Mode   int
}

func (d CommandFilterModeUpdate) commandString() string {
	return "FilterModeUpdate"
}

type CommandFilterCaseSensitiveUpdate struct {
	Filter        Filter
	CaseSensitive bool
}

func (d CommandFilterCaseSensitiveUpdate) commandString() string {
	return "FilterCaseSensitiveUpdate"
}

type CommandFilterKeyUpdate struct {
	Filter Filter
	Key    string
}

func (d CommandFilterKeyUpdate) commandString() string {
	return "FilterKeyUpdate"
}

type CommandToggleFollowMode struct {
}

func (d CommandToggleFollowMode) commandString() string {
	return "ToggleFollowMode"
}
