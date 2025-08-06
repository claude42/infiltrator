package model

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/claude42/infiltrator/config"
)

var (
	once        sync.Once
	busyCounter int
	busyState   atomic.Int32
)

type BusyState int

const (
	Idle = iota
	Busy
	BusyIdle
)

func BusySpin() {
	once.Do(func() {
		go updateBusySpinner()
	})

	// no need to do an atomic.Int32.Store() so frequently
	busyCounter++
	if busyCounter < 100000 {
		return
	}

	busyCounter = 0
	busyState.Store(Busy)
}

func updateBusySpinner() {
	cfg := config.GetConfiguration()
	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			switch busyState.Load() {
			case Busy:
				cfg.PostEventFunc(NewEventBusySpinnerUpdate(Busy))
				busyState.Store(BusyIdle)
			case BusyIdle:
				busyState.Store(Idle)
				cfg.PostEventFunc(NewEventBusySpinnerUpdate(Idle))
			case Idle:
				// nothing
			}
		case <-cfg.Context.Done():
			return
		}
	}
}
