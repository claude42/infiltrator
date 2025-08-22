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

	err = instance.kFormats.Load(file.Provider(formatsFile), toml.Parser())
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return
	}
	fail.OnError(err, "Loading formats file failed")

	err = instance.kFormats.Unmarshal("", &instance.formats)
	fail.OnError(err, "Error unmarshalling formats file")
}
