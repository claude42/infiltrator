package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/adrg/xdg"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/util"
	"github.com/spf13/pflag"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

var (
	PostEventFunc func(ev util.Event) error
	instance      *ConfigManager
)

type ConfigManager struct {
	userConfig *UserConfig  `koanf:"main"`
	panels     []PanelTable `koanf:"panel"`

	kConfig *koanf.Koanf `koanf:"-"`

	kFormats *koanf.Koanf      `koanf:"-"`
	formats  map[string]string `koanf:"-"`

	kState    *koanf.Koanf        `koanf:"-"`
	histories map[string][]string `koanf:"-"`
}

type UserConfig struct {
	Name     string `koanf:"name"`
	FileName string `koanf:"filename"`
	FilePath string `koanf:"filepath"`
	Stdin    bool   `koanf:"stdin"`
	Follow   bool   `koanf:"follow"`
	Lines    bool   `koanf:"lines"`
	Colorize bool   `koanf:"colorize"`
	Preset   string `koanf:"preset"`
	Debug    bool   `koanf:"debug"`

	FileFormat      string         `koanf:"-"`
	FileFormatRegex *regexp.Regexp `koanf:"-"`
}

type PanelTable struct {
	Type          string `koanf:"type"`
	Key           string `koanf:"key"`
	Mode          string `koanf:"mode"`
	CaseSensitive bool   `koanf:"casesensitive"`
	From          string `koanf:"from"`
	To            string `koanf:"to"`
}

type mainTable struct {
	Name     string `koanf:"name"`
	FileName string `koanf:"filename"`
	FilePath string `koanf:"filepath"`
	Stdin    bool   `koanf:"stdin"`
	Follow   bool   `koanf:"follow"`
	Lines    bool   `koanf:"lines"`
	Colorize bool   `koanf:"colorize"`
	Preset   string `koanf:"preset"`
	Debug    bool   `koanf:"debug"`
}

func init() {
	instance = &ConfigManager{
		kConfig:    koanf.New("."),
		kFormats:   koanf.New("."),
		kState:     koanf.New("."),
		formats:    make(map[string]string),
		histories:  make(map[string][]string),
		userConfig: &UserConfig{},
		panels:     make([]PanelTable, 0),
	}
}

func UserCfg() *UserConfig {
	return instance.userConfig
}

func Load() error {
	readDefaults(instance.kConfig)

	err := readConfigFile(instance.kConfig, mainConfigFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	err = readCommandLine(instance.kConfig)
	if err != nil {
		return err
	}

	if preset := instance.kConfig.String("main.preset"); preset != "" {
		err = readConfigFile(instance.kConfig, presetDir+preset+".toml")
		if err != nil {
			return err
		}
	}

	doSomeAdjustments()

	unmarshal()

	readStateFile()
	readFormatsFile()

	return nil
}

func doSomeAdjustments() {
	instance.kConfig.Set("main.filename", filepath.Base(instance.kConfig.String("main.filepath")))
}

func unmarshal() {

	foo := struct {
		main mainTable `koanf:"main"`
	}{}

	err := instance.kConfig.Unmarshal("", &foo)
	fail.OnError(err, "Unmarshalling failed failed")
}

func readDefaults(k *koanf.Koanf) {
	err := k.Load(confmap.Provider(defaults, "."), nil)
	fail.OnError(err, "Can't read defaults")
}

func readConfigFile(k *koanf.Koanf, lastPathPart string) error {
	path, err := xdg.ConfigFile(appName + lastPathPart)
	fail.OnError(err, "Can't determine preset filename")

	err = k.Load(file.Provider(path), toml.Parser())
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return err
	}
	fail.OnError(err, "Loading formats file failed")

	return nil
}

func readCommandLine(k *koanf.Koanf) error {
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

	switch len(flagSet.Args()) {
	case 0:
		instance.kConfig.Set("main.filename", "[stdin]")
		instance.kConfig.Set("main.filepath", "")
		instance.kConfig.Set("main.stdin", true)
	case 1:
		instance.kConfig.Set("main.filename", filepath.Base(flagSet.Args()[0]))
		instance.kConfig.Set("main.filepath", flagSet.Args()[0])
		instance.kConfig.Set("main.stdin", false)
	default:
		flagSet.Usage()
		return fmt.Errorf("try again")
	}

	return nil
}

func BuildFullPresetPath(presetName string) (fullPath string) {
	fullPath, err := xdg.ConfigFile(appName + presetDir + presetName + ".toml")
	fail.OnError(err, "Can't build path name!")

	return
}

func WritePreset(fullPath string) error {
	presetK := koanf.New(".")
	presetK.Load(structs.Provider(instance, "koanf"), nil)
	// presetK.Load(confmap.Provider(instance.kConfig.All(), "."), nil)

	presetK.Delete("main.filename")
	presetK.Delete("main.stdin")
	presetK.Delete("main.preset")
	presetK.Delete("main.debug")

	marshalledBytes, err := presetK.Marshal(toml.Parser())
	fail.OnError(err, "Error creating preset")

	err = os.MkdirAll(xdg.ConfigHome+"/"+appName+presetDir, 0755)
	fail.OnError(err, "Can't create preset directory")

	err = os.WriteFile(fullPath, marshalledBytes, 0644)
	fail.OnError(err, "Error creating preset")

	// return err
	return nil
}

func Formats() map[string]string {
	return instance.formats
}

func Panels() []PanelTable {
	return instance.panels
}

func SetPanels(panels []PanelTable) {
	instance.panels = panels
}
