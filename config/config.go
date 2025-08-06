package config

import (
	"context"
	"sync"

	"github.com/claude42/infiltrator/util"
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

	PostEventFunc func(ev util.Event) error
}

func GetConfiguration() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{}
	})
	return instance
}
