package sub

import (
	"errors"
	"testing"
)

func TestSubscriptionResult(t *testing.T) {
	sr := &SourceResponse{}
	s := Subscribe(Source(
		func() SourceResponse {
			sr.Result = 5
			return *sr
		}))
	r := <-s.Updates()
	err := s.Close()
	t.Logf("rcvd update=%+v err=%v\n", r, err)

	if err != nil {
		t.Errorf("Close() got error non-nil (%s), want nil", err)
	}
	if r != sr.Result {
		t.Errorf("Updates() channel value was %v, want %v", r, sr.Result)
	}
}

func TestSubscriptionError(t *testing.T) {
	sr := &SourceResponse{}
	s := Subscribe(Source(
		func() SourceResponse {
			sr.Err = errors.New("an error occurred")
			return *sr
		}))
	r := <-s.Updates()
	err := s.Close()
	t.Logf("rcvd update=%+v err=%v\n", r, err)

	if r != nil {
		t.Errorf("Updates() channel value was %v, want %v", r, sr.Result)
	}
	// Now we've done something at least once, close the subscription and read
	// its error to check it is set
	if err == nil {
		t.Error("Close() got error nil, want non-nil")
	}
}

func TestSubscriptionNext(t *testing.T) {
	exp := []string{"1", "2", "3"}
	sr := &SourceResponse{}
	s := Subscribe(Source(
		func() SourceResponse {
			sr.Result = exp
			return *sr
		}))
	r := <-s.Updates()
	err := s.Close()
	t.Logf("rcvd update=%+v err=%v\n", r, err)

	c := r.([]string)
	if c[0] != exp[0] || c[1] != exp[1] || c[2] != exp[2] {
		t.Errorf("Updates() channel value was %v, want %v", r, exp)
	}
	if err != nil {
		t.Errorf("Close() got error non-nil (%s), want nil", err)
	}
}
