package config

import (
	"context"
	"sync"
)

var (
	instance *ConfigManager
	once     sync.Once
)

type ConfigManager struct {
	FileName        string
	FilePath        string
	Stdin           bool
	ShowLineNumbers bool
	FollowFile      bool
	Debug           bool

	Quit chan string

	Context   context.Context
	Cancel    context.CancelFunc
	WaitGroup sync.WaitGroup
}

func GetConfiguration() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{}
	})
	return instance
}
