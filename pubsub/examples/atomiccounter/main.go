package main

import (
	"fmt"
	"time"

	"github.com/Emyrk/pubsub/publishers"
	"github.com/Emyrk/pubsub/pubregistry"
	"github.com/Emyrk/pubsub/subscribers"
)

// Atomic counter keeps all subscribers on the same level

func main() {
	reg := pubregistry.NewRegistry()
	pub := new(publishers.Base)
	panicError(reg.Register("/source", pub))

	for i := 0; i < 5; i++ {
		ValueWatcher(i, reg)
	}

	var i int64
	for {
		time.Sleep(1 * time.Second)
		fmt.Printf("Writing %d\n", i)
		pub.Write(i)
		i++
	}
}

func ValueWatcher(worker int, reg *pubregistry.Registry) {
	// Channel based subscription is just a channel written to by a publisher
	sub := subscribers.NewAtomicValueSubscriber()

	// Let's add callbacks
	callbackSub := subscribers.NewCallback(sub)
	callbackSub.AfterWrite = func(o interface{}) {
		if i, ok := o.(int64); ok {
			fmt.Printf("\t%d updated to %d\n", worker, i)
		}
	}

	panicError(reg.SubscribeTo("/source", callbackSub))
}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}
