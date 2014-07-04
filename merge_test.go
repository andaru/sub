package sub

import (
	"errors"
	"time"
	"testing"
)

func EchoIntSource(t *testing.T, i int) Sourcer {
	f := func() (sr SourceResponse) {
		sr.Result = i
		return
	}
	return Source(f)
}

func EchoIntNextSource(t *testing.T, i int) Sourcer {
	count := 0
	f := func() (sr SourceResponse) {
		sr.Result = i
		count++
		if count < 2 {
			sr.Next = time.Now()
		}
		return
	}
	return Source(f)
}

func EchoIntErrorSource(t *testing.T, i int) Sourcer {
	count := 0
	f := func() (sr SourceResponse) {
		sr.Result = i
		count++
		if count < 2 {
			sr.Next = time.Now()
		} else {
			sr.Err = errors.New("error")
		}
		return
	}
	return Source(f)
}

func TestMerge(t *testing.T) {
	// Create some subscribers and merge their results onto
	// a single Subcriber channel, receive results on that
	// single channel and perform comparison of the results
	n := 10
	var subs []Subscriber
	var err error

	expTotal := 0
	for i := 0; i < n; i++ {
		expTotal += i
		subs = append(subs, Subscribe(EchoIntSource(t, i)))
	}

	total := 0
	m := Merge(subs...)

	for {
		if total >= expTotal { err = m.Close(); break }
		select {
		case r := <-m.Updates():
			switch v := r.(type) {
			case nil: break
			case int: {
				total += v
			}
			}
		}
	}

	if total != expTotal {
		t.Errorf("got total [0..%d] %d, want %d",
			n, total, expTotal)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	t.Logf("n=%d total=%d expected=%d\n", n, total, expTotal)
}

func TestMergeWithNext(t *testing.T) {
	// create some subscribers that will re-fetch immediately,
	// stop after receiving twice as many responses
	n := 10
	var subs []Subscriber

	expTotal := 0
	for i := 0; i < n; i++ {
		expTotal += i
		subs = append(subs, Subscribe(EchoIntNextSource(t, i)))
	}

	expTotal = expTotal * 2
	m := Merge(subs...)

	var err error
	total := 0

	for {
		if total >= expTotal { err = m.Close(); break }
		select {
		case r := <-m.Updates():
			switch v := r.(type) {
			case nil: break
			case int: {
				total += v
			}
			}
		}
	}
	if total != expTotal {
		t.Errorf("got total [0..%d] %d, want %d",
			n, total, expTotal)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	t.Logf("n=%d total=%d expected=%d\n", n, total, expTotal)
}

func TestMergeWithError(t *testing.T) {
	// create some subscribers that will re-fetch immediately,
	// on the third time returning an error
	n := 10
	var subs []Subscriber

	expTotal := 0
	for i := 0; i < n; i++ {
		expTotal += i
		subs = append(subs, Subscribe(EchoIntErrorSource(t, i)))
	}

	expTotal = expTotal * 2
	m := Merge(subs...)

	var err error
	total := 0

	for {
		if total >= expTotal { err = m.Close(); break }
		select {
		case r := <-m.Updates():
			switch v := r.(type) {
			case nil: break
			case int: {
				total += v
			}
			}
		}
	}
	if total != expTotal {
		t.Errorf("got total [0..%d] %d, want %d",
			n, total, expTotal)
	}
	if err == nil {
		t.Errorf("expected an error here")
	}
	t.Logf("n=%d total=%d expected=%d\n", n, total, expTotal)
}
