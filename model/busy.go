package model

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/claude42/infiltrator/config"
)

var (
	once           sync.Once
	busyCounter    int
	busyState      atomic.Int32
	busyPercentage atomic.Int32
)

type BusyState int

const (
	Idle = iota
	Busy
	BusyIdle
)

func BusySpinPercentage(percentage int) {
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
	busyPercentage.Store(int32(percentage))
}

func BusySpin() {
	BusySpinPercentage(-1)
}

func updateBusySpinner() {
	cfg := config.GetConfiguration()
	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			switch busyState.Load() {
			case Busy:
				cfg.PostEventFunc(NewEventBusySpinnerUpdate(Busy, int(busyPercentage.Load())))
				busyState.Store(BusyIdle)
			case BusyIdle:
				busyState.Store(Idle)
				cfg.PostEventFunc(NewEventBusySpinnerUpdate(Idle, -1))
			case Idle:
				// nothing
			}
		case <-cfg.Context.Done():
			return
		}
	}
}
