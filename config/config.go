package config

import (
	"sync"

	"github.com/claude42/infiltrator/util"

	"github.com/knadh/koanf"
)

const (
	appName = "infiltrator"
)

var (
	instance *ConfigManager
	once     sync.Once
)

type ConfigManager struct {
	kState          *koanf.Koanf
	FileName        string
	FilePath        string
	Stdin           bool
	ShowLineNumbers bool
	FollowFile      bool
	Debug           bool

	PostEventFunc func(ev util.Event) error

	histories map[string][]string
}

func GetConfiguration() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{}
		instance.histories = make(map[string][]string)
		instance.kState = koanf.New(".")
	})
	return instance
}
