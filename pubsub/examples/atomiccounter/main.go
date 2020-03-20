package main

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/pubsub"
)

// Atomic counter keeps all subscribers on the same level

func main() {
	pub := pubsub.PubFactory.Base().Publish("/source")
	for i := 0; i < 5; i++ {
		ValueWatcher(i)
	}

	var i int64
	for {
		time.Sleep(1 * time.Second)
		fmt.Printf("Writing %d\n", i)
		pub.Write(i)
		i++
	}
}

func ValueWatcher(worker int) {
	// Channel based subscription is just a channel written to by a publisher
	pubsub.SubFactory.Value().Subscribe("/source", pubsub.SubCallbackWrap(nil,
		// Add a function call after every write
		func(o interface{}) {
			if i, ok := o.(int64); ok {
				fmt.Printf("\t%d updated to %d\n", worker, i)
			}
		}))
}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}
