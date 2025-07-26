package util

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type Observable interface {
	Watch(eh tcell.EventHandler)
	Unwatch(eh tcell.EventHandler)
}

type ObservableImpl struct {
	sync.Mutex
	watchers map[tcell.EventHandler]struct{}
}

func (o *ObservableImpl) Watch(eh tcell.EventHandler) {
	o.Lock()
	defer o.Unlock()

	if o.watchers == nil {
		o.watchers = make(map[tcell.EventHandler]struct{})
	}

	o.watchers[eh] = struct{}{}
}

func (o *ObservableImpl) Unwatch(eh tcell.EventHandler) {
	o.Lock()
	defer o.Unlock()

	if o.watchers == nil {
		return
	}

	delete(o.watchers, eh)

}

func (o *ObservableImpl) PostEvent(ev tcell.Event) {
	o.Lock()
	defer o.Unlock()

	if o.watchers == nil {
		return
	}

	watcherCopy := make(map[tcell.EventHandler]struct{}, len(o.watchers))
	for k := range o.watchers {
		watcherCopy[k] = struct{}{}
	}

	for k := range watcherCopy {
		k.HandleEvent(ev)
	}
}
