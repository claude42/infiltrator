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
	cm            *ConfigManager
)

type ConfigManager struct {
	UserConfig *UserConfig  `koanf:"main"`
	Panels     []PanelTable `koanf:"panel"`

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

func init() {
	cm = &ConfigManager{
		kConfig:    koanf.New("."),
		kFormats:   koanf.New("."),
		kState:     koanf.New("."),
		formats:    make(map[string]string),
		histories:  make(map[string][]string),
		UserConfig: &UserConfig{},
		Panels:     make([]PanelTable, 0),
	}
}

func User() *UserConfig {
	return cm.UserConfig
}

func Formats() map[string]string {
	return cm.formats
}

func Panels() []PanelTable {
	return cm.Panels
}

func SetPanels(panels []PanelTable) {
	cm.Panels = panels
}

func Load() error {
	readDefaults()

	err := readConfigFile(mainConfigFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	err = readCommandLine()
	if err != nil {
		return err
	}

	if preset := cm.kConfig.String("main.preset"); preset != "" {
		err = readConfigFile(presetDir + preset + ".toml")
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
	cm.kConfig.Set("main.filename", filepath.Base(cm.kConfig.String("main.filepath")))
}

func unmarshal() {
	err := cm.kConfig.Unmarshal("", &cm)
	fail.OnError(err, "Unmarshalling failed failed")
}

func readDefaults() {
	err := cm.kConfig.Load(confmap.Provider(defaults, "."), nil)
	fail.OnError(err, "Can't read defaults")
}

func readConfigFile(lastPathPart string) error {
	path, err := xdg.ConfigFile(appName + lastPathPart)
	fail.OnError(err, "Can't determine preset filename")

	err = cm.kConfig.Load(file.Provider(path), toml.Parser())
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return err
	}
	fail.OnError(err, "Loading formats file failed")

	return nil
}

func readCommandLine() error {
	flagSet := pflag.NewFlagSet("infilt", pflag.ContinueOnError)
	lines := flagSet.BoolP("lines", "l", false, "Show line numbers")
	follow := flagSet.BoolP("follow", "f", false, "Follow changes to file")
	colorize := flagSet.BoolP("colorize", "c", true, "Colorize output if it's in a well known format")
	debug := flagSet.BoolP("debug", "d", false, "Log debugging information to ./debug.log")
	preset := flagSet.StringP("preset", "p", "", "Load preset by name")

	err := flagSet.Parse(os.Args[1:])
	fail.OnError(err, "Parsing of command line failed")

	if flagSet.Lookup("lines").Changed {
		err := cm.kConfig.Set("main.lines", *lines)
		fail.OnError(err, "Error setting command line option")
	}

	if flagSet.Lookup("follow").Changed {
		err := cm.kConfig.Set("main.follow", *follow)
		fail.OnError(err, "Error setting command line option")
	}

	if flagSet.Lookup("colorize").Changed {
		err := cm.kConfig.Set("main.colorize", *colorize)
		fail.OnError(err, "Error setting command line option")
	}

	if flagSet.Lookup("debug").Changed {
		err := cm.kConfig.Set("main.debug", *debug)
		fail.OnError(err, "Error setting command line option")
	}

	if flagSet.Lookup("preset").Changed {
		err := cm.kConfig.Set("main.preset", *preset)
		fail.OnError(err, "Error setting command line option")
	}

	switch len(flagSet.Args()) {
	case 0:
		cm.kConfig.Set("main.filename", "[stdin]")
		cm.kConfig.Set("main.filepath", "")
		cm.kConfig.Set("main.stdin", true)
	case 1:
		cm.kConfig.Set("main.filename", filepath.Base(flagSet.Args()[0]))
		cm.kConfig.Set("main.filepath", flagSet.Args()[0])
		cm.kConfig.Set("main.stdin", false)
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
	presetK.Load(structs.Provider(cm, "koanf"), nil)
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
