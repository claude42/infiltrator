package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/adrg/xdg"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/util"
	"github.com/spf13/pflag"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

var (
	instance *ConfigManager
	once     sync.Once
)

type ConfigManager struct {
	FileName        string
	FilePath        string
	FileFormat      string
	FileFormatRegex *regexp.Regexp
	Stdin           bool

	PostEventFunc func(ev util.Event) error

	kState    *koanf.Koanf
	histories map[string][]string

	kFormats *koanf.Koanf
	Formats  map[string]string

	kConfig    *koanf.Koanf
	UserConfig ConfigFile
}

type MainTable struct {
	Name     string `koanf:"name"`
	FileName string `koanf:"filename"`
	Follow   bool   `koanf:"follow"`
	Lines    bool   `koanf:"lines"`
	Colorize bool   `koanf:"colorize"`
	Preset   string `koanf:"preset"`
	Debug    bool   `koanf:"debug"`
}

type PanelTable struct {
	Type          string `koanf:"type"`
	Key           string `koanf:"key"`
	Mode          string `koanf:"mode"`
	CaseSensitive bool   `koanf:"casesensitive"`
	From          string `koanf:"from"`
	To            string `koanf:"to"`
}

type ConfigFile struct {
	Main   MainTable    `koanf:"main"`
	Panels []PanelTable `koanf:"panel"`
}

func GetConfiguration() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{}
		instance.histories = make(map[string][]string)
		instance.kState = koanf.New(".")
		instance.kFormats = koanf.New(".")
		instance.kConfig = koanf.New(".")
	})
	return instance
}

func (cm *ConfigManager) Load() error {
	cm.ReadDefaults(cm.kConfig)

	err := cm.ReadConfigFile(cm.kConfig, mainConfigFileName)
	if err != nil {
		return err
	}

	err = cm.ReadCommandLine(cm.kConfig)
	if err != nil {
		return err
	}

	if cm.UserConfig.Main.Preset != "" {
		err = cm.ReadConfigFile(cm.kConfig, presetDir+cm.UserConfig.Main.Preset+".toml")
		if err != nil {
			return err
		}
	}
	cm.ReadStateFile()
	cm.ReadFormatsFile()

	return nil
}

func (cm *ConfigManager) ReadDefaults(k *koanf.Koanf) {
	err := k.Load(structs.Provider(defaults, "koanf"), nil)
	fail.OnError(err, "Can't read defaults")

	err = k.Unmarshal("", &cm.UserConfig)
	fail.OnError(err, "Unmarshalling failed failed")
}

func (cm *ConfigManager) ReadConfigFile(k *koanf.Koanf, lastPathPart string) error {
	path, err := xdg.ConfigFile(appName + lastPathPart)
	fail.OnError(err, "Can't determine preset filename")

	err = k.Load(file.Provider(path), toml.Parser())
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return err
	}
	fail.OnError(err, "Loading formats file failed")

	err = k.Unmarshal("", &cm.UserConfig)
	fail.OnError(err, "Unmarshalling failed failed")

	return nil
}

func (cm *ConfigManager) ReadCommandLine(k *koanf.Koanf) error {
	flagSet := pflag.NewFlagSet("infilt", pflag.ContinueOnError)
	lines := flagSet.BoolP("lines", "l", false, "Show line numbers")
	follow := flagSet.BoolP("follow", "f", false, "Follow changes to file")
	colorize := flagSet.BoolP("colorize", "c", true, "Colorize output if it's in a well known format")
	debug := flagSet.BoolP("debug", "d", false, "Log debugging information to ./debug.log")
	preset := flagSet.StringP("preset", "p", "", "Load preset by name")

	err := flagSet.Parse(os.Args[1:])
	fail.OnError(err, "Parsing of command line failed")

	if flagSet.Lookup("lines").Changed {
		err := k.Set("main.lines", *lines)
		fail.OnError(err, "Error setting command line option")
	}

	if flagSet.Lookup("follow").Changed {
		err := k.Set("main.follow", *follow)
		fail.OnError(err, "Error setting command line option")
	}

	if flagSet.Lookup("colorize").Changed {
		err := k.Set("main.colorize", *colorize)
		fail.OnError(err, "Error setting command line option")
	}

	if flagSet.Lookup("debug").Changed {
		err := k.Set("main.debug", *debug)
		fail.OnError(err, "Error setting command line option")
	}

	if flagSet.Lookup("preset").Changed {
		err := k.Set("main.preset", *preset)
		fail.OnError(err, "Error setting command line option")
	}

	err = k.Unmarshal("", &cm.UserConfig)
	fail.OnError(err, "Unmarshalling failed failed")

	switch len(flagSet.Args()) {
	case 0:
		cm.FileName = "[stdin]"
		cm.FilePath = ""
		cm.Stdin = true
	case 1:
		cm.FilePath = flagSet.Args()[0]
		cm.FileName = filepath.Base(cm.FilePath)
		cm.Stdin = false
	default:
		flagSet.Usage()
		return fmt.Errorf("try again")
	}

	return nil
}

// Untested as of now
func (cm *ConfigManager) WritePreset(k *koanf.Koanf, lastPathPart string) error {
	var presetK = koanf.New(".")

	err := k.Cut("panel").Merge(presetK)
	fail.OnError(err, "Error creating preset")

	err = presetK.Set("main.filename", k.Bool("main.filename"))
	fail.OnError(err, "Error creating preset")

	err = presetK.Set("main.follow", k.Bool("main.follow"))
	fail.OnError(err, "Error creating preset")

	err = presetK.Set("main.lines", k.Bool("main.lines"))
	fail.OnError(err, "Error creating preset")

	err = presetK.Set("main.colorize", k.Bool("main.colorize"))
	fail.OnError(err, "Error creating preset")

	marshalledBytes, err := presetK.Marshal(toml.Parser())
	fail.OnError(err, "Error creating preset")

	var path string
	path, err = xdg.ConfigFile(appName + lastPathPart)
	fail.OnError(err, "Can't determine preset filename")

	err = os.WriteFile(path, marshalledBytes, 0644)

	return err
}
