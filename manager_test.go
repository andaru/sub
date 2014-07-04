package sub

import (
	"testing"
	"time"
)

func getFetcher() SourceFunc {
	i := 0
	return func() (sr SourceResponse) {
		sr.Err = nil
		sr.Result = i
		sr.Next = time.Now()
		i++
		return
	}
}

func TestSetGetDel(t *testing.T) {
	m := NewManager()
	m.Set("test1", Subscribe(Source(getFetcher())))
	r1 := m.Get("test1")
	r1.(Subscriber).Close()
	// We should get one result if we close the subscription,
	// since the updates channel is buffered.

	if r1 == nil {
		t.Error("test1 key nil, want non-nil")
	}
	r2 := m.Get("test2")
	if r2 != nil {
		t.Errorf("test2 key %v, want nil", r2)
	}
	m.Delete("test1")
	r1 = m.Get("test1")
	if r1 != nil {
		t.Error("test1 not deleted, want deleted")
	}
}

func TestGetSubscribers(t *testing.T) {
	m := NewManager()
	m.Set("test1", Subscribe(Source(getFetcher())))
	m.Set("test2", Subscribe(Source(getFetcher())))
	m.Set("test3", Subscribe(Source(getFetcher())))

	subs := m.GetSubscribers()
	if len(subs) != 3 {
		t.Errorf("Expected 3 subscriptions, got %d", len(subs))
	}
}

func TestGetSubscriberMap(t *testing.T) {
	m := NewManager()
	m.Set("test1", Subscribe(Source(getFetcher())))
	m.Set("test2", Subscribe(Source(getFetcher())))
	m.Set("test3", Subscribe(Source(getFetcher())))

	subs := m.GetSubscriberMap()
	if len(subs) != 3 {
		t.Errorf("Expected 3 subscriptions, got %d", len(subs))
	}
	if subs["test1"] == nil {
		t.Error("expected test1 key to be valid, was not")
	}
	if subs["test2"] == nil {
		t.Error("expected test1 key to be valid, was not")
	}
	if subs["test3"] == nil {
		t.Error("expected test1 key to be valid, was not")
	}
}

func TestSubscriptionInManager(t *testing.T) {
	m := NewManager()
	m.Set("test1", Subscribe(Source(getFetcher())))

	r1 := m.Get("test1")
	// Test we can do something with the subscription, i.e., we get a result
	// within the timeout
	period := 5
	timeout := time.After(time.Duration(period) * time.Second)

	n := 42
	select {
	case result := <-r1.(Subscriber).Updates():
		v := result.(int)
		if v >= n {
			r1.(Subscriber).Close()
			break
		}
	case <-timeout:
		t.Error("Timed out waiting for infinite sequence source")
		return
	}
	t.Logf("Successfully received %d updates within %ds", n, period)
}
