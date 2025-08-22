package busy

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/util"
)

var (
	busyCounter    int
	busyState      atomic.Int32
	busyPercentage atomic.Int32
)

func SpinWithPercentage(percentage int) {
	// no need to do an atomic.Int32.Store() so frequently
	busyCounter++
	if busyCounter < 100 {
		return
	}

	busyCounter = 0
	busyState.Store(Busy)
	busyPercentage.Store(int32(percentage))
}

func SpinWithFraction(a, b int) {
	if b == 0 {
		Spin()
	} else {
		SpinWithPercentage(100 * a / b)
	}
}

func Spin() {
	SpinWithPercentage(-1)
}

func StartBusySpinner(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			switch busyState.Load() {
			case Busy:
				config.PostEventFunc(NewEventBusySpinnerUpdate(Busy, int(busyPercentage.Load())))
				busyState.Store(BusyIdle)
			case BusyIdle:
				busyState.Store(Idle)
				config.PostEventFunc(NewEventBusySpinnerUpdate(Idle, -1))
			case Idle:
				// nothing
			}
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

type State int

const (
	Idle = iota
	Busy
	BusyIdle
)

// created by busy

type EventBusySpinnerUpdate struct {
	util.EventImpl

	BusyState      State
	BusyPercentage int
}

func NewEventBusySpinnerUpdate(busyState State, percentage int) *EventBusySpinnerUpdate {
	ev := &EventBusySpinnerUpdate{BusyState: busyState, BusyPercentage: percentage}
	ev.EventImpl.SetWhen()
	return ev
}
