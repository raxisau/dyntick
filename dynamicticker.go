package dyntick

import (
	"sync"
	"time"
)

// DynamicTicker The guts of this
type DynamicTicker struct {
	C       chan time.Time
	stop    chan struct{}
	stopped bool
	id      int
	lock    *sync.Mutex
	tt      *time.Ticker
}

// NewDynamicTicker create a new dynamic ticker
func NewDynamicTicker(d time.Duration) *DynamicTicker {

	t := &DynamicTicker{
		C:       make(chan time.Time, 1),
		stop:    make(chan struct{}, 1),
		tt:      time.NewTicker(d),
		stopped: false,
		id:      1,
		lock:    &sync.Mutex{},
	}
	go t.tickProducer(t.id)

	return t
}

func (t *DynamicTicker) tickProducer(id int) {
	localTicker := t.tt
	defer localTicker.Stop()

	for !t.stopped {
		select {
		case now := <-localTicker.C:

			// If the chanel is not ready then just drop this tick
			select {
			case t.C <- now:
			default:
				// discard the tick if you are not ready to take it
			}
		case <-t.stop:
			return
		}
	}
}

// Stop Stops the dynamic ticker
func (t *DynamicTicker) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.stopped {
		return
	}

	t.stopped = true
	t.stop <- struct{}{}
}

// Reset changes the duration
func (t *DynamicTicker) Reset(d time.Duration) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.stopped {
		return
	}

	// Create a new ticket and a dynamic ticker producer
	t.tt = time.NewTicker(d)
	t.id++

	// stop the old ticker
	t.stop <- struct{}{}

	go t.tickProducer(t.id)
}
