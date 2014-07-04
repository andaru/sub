package sub

import (
	"log"
	"math"
)

type Publisher interface {
	// Register and unregister new listeners; the subscriber
	// returned will receive its' Updates() from the Publisher
	Subscribe() Subscriber
	Unsubscribe(...Subscriber)
	Close() error
	// Start listening to the data source
	StartSource()
	StopSource()
	// Submit some data to all listeners
	// Publish(interface{})
}

// This basic publisher takes a single source subscription and makes
// it available over the Publisher interface by calling its Publish
// interface for data items arriving from the source Subscriber,
// which will broadcast to all listeners.
type publisher struct {
	source Subscriber

	input chan interface{}
	sub   chan *fanoutSubscriber
	unsub chan *fanoutSubscriber

	start chan struct{}
	stop  chan struct{}
	quit  chan struct{}
	errs  chan error

	listeners map[*fanoutSubscriber]bool
}

func min(lhs, rhs int) int {
	return int(math.Min(float64(lhs), float64(rhs)))
}

func max(lhs, rhs int) int {
	return int(math.Max(float64(lhs), float64(rhs)))
}

func NewPublisher(source Subscriber, queuelen int) Publisher {
	// Use a safe minimum queue length bounds to avoid deadlock
	// with aggressive sources. The go runtime tends to be able to
	// keep up with feeding any number of subscribers with queue
	// depth of 2, sometimes 3.  Use a larger number for a little
	// more safety.

	// At least one Subscriber should be gotten from the Publisher before
	// calling StartSource().
	qlen := max(queuelen, 10)
	log.Printf("qlen=%d\n", qlen)
	p := &publisher{
		source:    source,
		input:     make(chan interface{}, qlen),
		sub:       make(chan *fanoutSubscriber),
		unsub:     make(chan *fanoutSubscriber),
		start:     make(chan struct{}),
		stop:      make(chan struct{}),
		quit:      make(chan struct{}),
		errs:      make(chan error),
		listeners: make(map[*fanoutSubscriber]bool),
	}
	go p.loop()
	return p
}

func (p *publisher) Close() (err error) {
	err = p.source.Close()
	log.Printf("%p Close() err=%v", p, err)
	return err
}

func (p *publisher) Subscribe() Subscriber {
	// attach a new listener and provide a subscriber interface to it
	sub := newFanout(cap(p.input))
	p.sub <- sub
	return sub
}

func (p *publisher) Unsubscribe(s ...Subscriber) {
	for _, sub := range s {
		p.unsub <- sub.(*fanoutSubscriber)
	}
}

func (p *publisher) Publish(data interface{}) {
	if data != nil {
		p.input <- data
	}
}

func (s *fanoutSubscriber) publish(data interface{}) {
	if data != nil {
		s.input <- data
	}
}

type fanoutSubscriber struct {
	input chan interface{}
}

func newFanout(queuelen int) *fanoutSubscriber {
	return &fanoutSubscriber{
		input: make(chan interface{}, queuelen),
	}
}

func (s *fanoutSubscriber) Updates() <-chan interface{} {
	return s.input
}

func (s *fanoutSubscriber) Close() error {
	return nil
}

func (p *publisher) StartSource() {
	p.start <- struct{}{}
}

func (p *publisher) StopSource() {
	p.stop <- struct{}{}
}

func (p *publisher) loop() {
	var updates <-chan interface{} = nil
	for {
		select {
		case <-p.start:
			// Start broadcasting updates from the Source
			log.Printf("[%p] <- p.start (source)\n", p)
			// enable the input channel (by setting it to not null)
			updates = p.source.Updates()
		case <-p.stop:
			// Stop broadcasting the Source's updates
			log.Printf("[%p] <- p.stop (source)\n", p)
			updates = nil
		case s, ok := <-p.sub:
			// Handles the addition of a new subscriber
			if ok {
				p.listeners[s] = true
			} else {
				return
			}
		case uns := <-p.unsub:
			// Handles removal of a subscriber
			log.Printf("<- p.unsub s=%+v\n", uns)
			delete(p.listeners, uns)
		case update := <-updates:
			// push onto the buffered input
			p.Publish(update)
		case pubval := <-p.input:
			// Values from the input (via the source)
			// publish to all listeners
			for sub := range p.listeners {
				sub.publish(pubval)
			}
		case <-p.quit:
			log.Println("<- p.quit")
			p.errs <- p.Close()
			return
		}
	}
}

//

//
