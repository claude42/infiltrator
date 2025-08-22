package config

import (
	"errors"
	"os"

	"github.com/adrg/xdg"
	"github.com/claude42/infiltrator/fail"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
)

const formatsFileName = "/formats.toml"

// TODO error handling
func readFormatsFile() {
	formatsFile, err := xdg.ConfigFile(appName + formatsFileName)
	fail.OnError(err, "Can't determine formats File")

	err = cm.kFormats.Load(file.Provider(formatsFile), toml.Parser())
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return
	}
	fail.OnError(err, "Loading formats file failed")

	err = cm.kFormats.Unmarshal("", &cm.formats)
	fail.OnError(err, "Error unmarshalling formats file")
}
