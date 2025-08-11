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

func (cm *ConfigManager) AddToHistory(filter string, value string) {
	filterHistory := cm.histories[filter]
	util.Remove(filterHistory, value)
	filterHistory = append([]string{value}, filterHistory[:min(len(filterHistory), 99)]...)
	cm.histories[filter] = filterHistory
}

func (cm *ConfigManager) FromHistory(filter string, index int) (string, error) {
	fail.If(cm == nil, "No config manager?!")
	fail.If(cm.histories == nil, "No histories?!")

	history, ok := cm.histories[filter]
	if !ok {
		return "", util.ErrNotFound
	}

	if index >= len(history) {
		return "", util.ErrNotFound
	}

	return history[index], nil
}

func (cm *ConfigManager) ReadStateFile() {
	stateFile, err := xdg.StateFile(appName + "/history.toml")
	fail.OnError(err, "Can't determine State File")

	err = cm.kState.Load(file.Provider(stateFile), toml.Parser())
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return
	}
	fail.OnError(err, "Loading state file failed")

	for _, historyName := range Histories {
		cm.histories[historyName] = cm.kState.Strings((historyName + ".history"))
	}

	// cm.keywordHistory = append(cm.keywordHistory, "hurahagl")

}

func (cm ConfigManager) WriteStateFile() error {
	stateFile, err := xdg.StateFile(appName + "/history.toml")
	fail.OnError(err, "Can't determine State File")

	for _, historyName := range Histories {
		err = cm.kState.Set(historyName+".history", cm.histories[historyName])
		fail.OnError(err, "Error storing new history lines")
	}

	marshalledBytes, err := cm.kState.Marshal(toml.Parser())
	fail.OnError(err, "Error marshalling data")

	err = os.WriteFile(stateFile, marshalledBytes, 0644)
	// TODO: handle error differently?

	return err
}
