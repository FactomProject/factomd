package main

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/pubsub"
)

// If you really want to make a subscriber outside the package, you can.
// Since the functions are unexported in the pubsub package, you must use a
// SubEmbedded
func main() {
	pub := new(pubsub.PubBase).Publish("/source")
	s := NewExternalSub()
	_ = pubsub.GlobalRegistry().SubscribeTo("/source", s)

	var i int64
	for {
		time.Sleep(1 * time.Second)
		fmt.Printf("Writing %d\n", i)
		pub.Write(i)
		i++
	}
}

type ExternalSub struct {
	pubsub.SubEmbedded
}

func NewExternalSub() *ExternalSub {
	s := new(ExternalSub)

	// Set the subscriber functions that the publisher will call
	s.Done = func() {}
	s.Write = func(o interface{}) {
		fmt.Println("A write occurred!")
	}
	s.SetUnsubscribe = func(ubsub func()) {}

	return s
}
