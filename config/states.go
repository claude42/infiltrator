package config

import (
	"errors"
	"os"

	"github.com/adrg/xdg"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/util"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
)

const historyFileName = "/history.toml"

func AddToHistory(filter string, value string) {
	filterHistory := instance.histories[filter]
	util.Remove(filterHistory, value)
	filterHistory = append([]string{value}, filterHistory[:min(len(filterHistory), 99)]...)
	instance.histories[filter] = filterHistory
}

func FromHistory(filter string, index int) (string, error) {
	fail.If(instance == nil, "No config manager?!")
	fail.If(instance.histories == nil, "No histories?!")

	history, ok := instance.histories[filter]
	if !ok {
		return "", util.ErrNotFound
	}

	if index >= len(history) {
		return "", util.ErrNotFound
	}

	return history[index], nil
}

func readStateFile() {
	stateFile, err := xdg.StateFile(appName + historyFileName)
	fail.OnError(err, "Can't determine State File")

	err = instance.kState.Load(file.Provider(stateFile), toml.Parser())
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return
	}
	fail.OnError(err, "Loading state file failed")

	for _, historyName := range Histories {
		instance.histories[historyName] = instance.kState.Strings((historyName + ".history"))
	}

	// cm.keywordHistory = append(cm.keywordHistory, "hurahagl")

}

func WriteStateFile() error {
	stateFile, err := xdg.StateFile(appName + "/history.toml")
	fail.OnError(err, "Can't determine State File")

	for _, historyName := range Histories {
		err = instance.kState.Set(historyName+".history", instance.histories[historyName])
		fail.OnError(err, "Error storing new history lines")
	}

	marshalledBytes, err := instance.kState.Marshal(toml.Parser())
	fail.OnError(err, "Error marshalling data")

	err = os.MkdirAll(xdg.StateHome+"/"+appName, 0755)
	fail.OnError(err, "Can't create state directory")

	err = os.WriteFile(stateFile, marshalledBytes, 0644)
	// TODO: handle error differently?

	return err
}
