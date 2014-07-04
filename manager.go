package sub

import (
	"sync"
)

// The manager looks after a collection of subscriptions of different
// types with different subscription functions and parameters.

type SubscriberManager interface {
	Get(key string) Subscriber
	GetSubscribers() []Subscriber
	GetSubscriberMap() map[string]Subscriber
	Set(key string, sub Subscriber)
	Delete(key string)
}

type Manager struct {
	sync.Mutex
	subs map[string]Subscriber
}

func NewManager() *Manager {
	m := &Manager{subs: make(map[string]Subscriber)}
	return m
}

func (m *Manager) Get(key string) Subscriber {
	m.Lock()
	defer m.Unlock()
	return m.subs[key]
}

func (m *Manager) GetSubscribers() []Subscriber {
	m.Lock()
	defer m.Unlock()
	subs := make([]Subscriber, len(m.subs))
	i := 0
	for key := range m.subs {
		subs[i] = m.subs[key]
	}
	return subs
}

func (m *Manager) GetSubscriberMap() map[string]Subscriber {
	m.Lock()
	defer m.Unlock()
	subs := make(map[string]Subscriber)
	for k := range m.subs {
		subs[k] = m.subs[k]
	}
	return subs
}

func (m *Manager) Set(key string, sub Subscriber) {
	m.Lock()
	defer m.Unlock()
	m.subs[key] = sub
}

func (m *Manager) Delete(key string) {
	m.Lock()
	defer m.Unlock()
	delete(m.subs, key)
}
