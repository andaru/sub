package sub

import (
	"log"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestPubsub(t *testing.T) {

	source := Subscribe(Source(
		func() SourceResponse {
			sr := SourceResponse{}
			sr.Result = rand.Intn(10)
			// Produce another result immediately
			sr.Next = time.Now()
			return sr
		}))
	log.Println("getting publisher")
	p := NewPublisher(source, 3)
	n := 100000
	log.Println("getting subscriptions..")
	subs := make([]Subscriber, n)
	for i := 0; i < n; i++ {
		subs[i] = p.Subscribe()
	}
	t.Logf("created Publisher with %d subscriptions\n", len(subs))

	p.StartSource()
	m := Merge(subs...)

	tot := 0
	for r := range m.Updates() {
		tot += r.(int)
		if tot > 2500000 {
			m.Close()
		}
	}
	p.Close()
}



//
