package sub

type merged struct {
	subs []Subscriber
	updates chan interface {}
	quit chan struct{}
	errs chan error
}

func (m *merged) Updates() <-chan interface{} {
	return m.updates
}

func (m *merged) Close() (err error) {
	close(m.quit)
	for _ = range m.subs {
		// XXX log the error?
		if e := <-m.errs; e != nil { err = e }
	}
	close(m.updates)
	return
}

// Merges the subscribers by combining all of their
// update and error channels into a single Subscriber
func Merge(subs ...Subscriber) Subscriber {
	m := &merged{
		subs: subs,
		updates: make(chan interface {}),
		quit: make(chan struct{}),
		errs: make(chan error),
	}

	// Merge subscriptions, errors and closing events
	for _, sub := range subs {
		go func(s Subscriber) {
			for {
				var update interface{}
				select {
				case update = <- s.Updates():
				case <-m.quit:
					m.errs <- s.Close()
					return
				}
				select {
				case m.updates <- update:
				case <-m.quit:
					m.errs <- s.Close()
					return
				}
			}
		}(sub)
	}
	return m
}
