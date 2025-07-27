package config

import (
	"sync"
)

var (
	instance *ConfigManager
	once     sync.Once
)

type ConfigManager struct {
	FileName        string
	ShowLineNumbers bool
	FollowFile      bool
}

func GetConfiguration() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{}
	})
	return instance
}
