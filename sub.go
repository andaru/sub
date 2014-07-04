package sub

import (
	"time"
)

const (
	DEFAULT_PERIOD = time.Hour
)

type Subscriber interface {
	Updates() <-chan interface{}
	Close() error
}

type Sourcer interface {
	Fetch() SourceResponse
}

// The result of a SourceFunc, includes the function return value
// (Result) any error and the next time this function should run.
type SourceResponse struct {
	Result interface{} // one or more results
	Next   time.Time   // the next time this subscriber should run
	Err    error
}

func newSourceResponse() *SourceResponse {
	return &SourceResponse{}
}

// A SourceFunc is the function which fetches the source data,
// returning it as Result in the SearchResponse, along with
// the Next time to fetch (the default value will never fetch
// again) and any error that occurred.
type SourceFunc func() SourceResponse

// The generic subscriber struct. Call Subscribe to subscribe to a
// Sourcer fetch function.
type subscriber struct {
	updates chan interface{}
	closing chan chan error
	source  Sourcer
	period  time.Duration // duration between start of each run
}

// Subscribes to the source immediately in a goroutine, returning the
// running subscriber. Receive subscription updates by receiving on
// the return value's Updates() method, e.g.,
//
// sub := subscribe.Subscribe(statsCollector)
//
// for { select {
//    case result := <-sub.Updates():
//    ...
// } }

func Subscribe(source Sourcer) Subscriber {
	s := &subscriber{
		updates: make(chan interface{}),
		closing: make(chan chan error),
		source:  source,
		period:  DEFAULT_PERIOD,
	}
	go s.loop()
	return s
}

func (self *subscriber) Close() error {
	errc := make(chan error)
	self.closing <- errc
	return <-errc
}

func (self *subscriber) Updates() <-chan interface{} {
	return self.updates
}

const (
	MaxInt64    = int64(^uint64(0) >> 1)
	MaxDuration = time.Duration(MaxInt64)
)

// The subscriber's main loop.

// See if we should be scheduled to run, and if we're ready to run,
// execute the Sourcer in a goroutine, and feed its result into
// a pending results queue, from which results are emitted to the
// channel available on Updates(). If we are Close()d, shut down
// all channels, emit any errors and clean up before returning.

func (self *subscriber) loop() {
	var done chan SourceResponse
	var pending []interface{}
	var next time.Time
	var err error

	// To schedule ourselves for the first run, set a time that's
	// not absolute zero, which we overload to mean don't schedule
	// again. This means a SourceResponse returned from a Sourcer
	// without a .next member explicitly set will not execute
	// again. In any event, the loop continues until Close() is called.
	next = time.Time(time.Unix(1, 0))

	for {
		var delay time.Duration

		var start <-chan time.Time
		// schedule a next start time
		if now := time.Now(); next.After(now) {
			delay = next.Sub(now)
		}

		if next.IsZero() {
			start = time.After(MaxDuration)
		} else if done == nil {
			start = time.After(delay)
		}

		var first interface{}
		var updates chan interface{}

		if len(pending) > 0 {
			first = pending[0]
			updates = self.updates
		}

		select {
		case <-start:
			done = make(chan SourceResponse, 1)
			go func() {
				done <- self.source.Fetch()
			}()
		case sr := <-done:
			done = nil // allow ourselves to be scheduled again
			err = sr.Err
			next = sr.Next
			pending = append(pending, sr.Result)
		case updates <- first:
			pending = pending[1:]
		case errc := <-self.closing:
			errc <- err
			close(self.updates)
			return
		}
	}
}

// A source is a Sourcer
type source struct {
	f SourceFunc
}

func newSource(f SourceFunc) Sourcer {
	return &source{f: f}
}

// Creates a source that runs the function with arugments
func Source(f SourceFunc) Sourcer {
	return newSource(f)
}

func (self *source) Fetch() SourceResponse {
	c := make(chan SourceResponse)
	go func() {
		c <- self.f()
	}()
	return <-c
}
