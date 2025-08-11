package util

import "time"

const defaultDelay = 300 * time.Millisecond

type Delay struct {
	invokeFunc func()
	duration   time.Duration
	ch         chan struct{}
}

func NewDelay(invokeFunc func()) *Delay {
	d := &Delay{
		invokeFunc: invokeFunc,
		duration:   defaultDelay,
	}

	d.ch = make(chan struct{}, 10)

	go d.run()

	return d
}

func NewCustomDelay(invokeFunc func(), duration time.Duration) *Delay {
	d := &Delay{
		invokeFunc: invokeFunc,
		duration:   duration,
	}

	d.ch = make(chan struct{}, 10)

	go d.run()

	return d
}

// func NewCustomDelay(invokeFunc func(), duration time.Duration) *Delay {
// 	d := NewDelay(invokeFunc)

// 	d.duration = duration

// 	return d
// }

func (d *Delay) SetInvokeFunc(invokeFunc func()) {
	d.invokeFunc = invokeFunc
}

func (d *Delay) SetDelay(duration time.Duration) {
	d.duration = duration
}

func (d *Delay) Now() {
	d.ch <- struct{}{}
}

func (d *Delay) run() {
	timer := time.NewTimer(d.duration)
	defer timer.Stop()

	for {
		select {
		case _, ok := <-d.ch:
			if !ok {
				// should never happen
				return
			}
			timer.Reset(d.duration)
		case <-timer.C:
			d.invokeFunc()
		}
	}
}

func (d *Delay) Stop() {
	close(d.ch)
}
