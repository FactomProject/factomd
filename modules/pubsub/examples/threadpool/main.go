package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/FactomProject/factomd/modules/pubsub"
)

const buffer int = 100

func main() {
	// This code finds the total number of prime numbers <= max
	max := int64(1e5)

	// One person publishes the work on a round robin basis
	robinPub := pubsub.PubFactory.RoundRobin(buffer).Publish("/source")
	go robinPub.Start() // The writer is threaded

	// Agg is a publisher with many writers. Need to publish so a subscriber
	// can find it.
	pubsub.PubFactory.Threaded(100).Publish("/aggregate", pubsub.PubMultiWrap())

	workers := 5
	for i := 0; i < workers; i++ {
		// Calculates if a number is prime. They will publish any primes
		// to the agg publisher. When the source publisher is done, and the
		// worker has no more tasks, it will close the agg publisher.
		// The agg publisher will close when all writers are done.
		go PrimeWorker()
	}

	// Load the work
	go func() {
		for robinPub.NumberOfSubscribers() != workers {
			time.Sleep(25 * time.Millisecond)
		}
		for i := int64(0); i < max; i++ {
			robinPub.Write(i) // Write all numbers to the /source
		}
		_ = robinPub.Close() // Close when all numbers are done being read
	}()

	// Print the results. Count will wait for agg to be closed, then print the
	// count of things published to agg.
	fmt.Printf("%d primes found\n", Count())
}

func PrimeWorker() {
	// Channel based subscription is just a channel written to by a publisher
	sub := pubsub.SubFactory.Channel(5).Subscribe("/source")

	// agg is where we write our results.
	//agg := pubsub.PubFactory.Multi(buffer).Publish("/aggregate")
	agg := pubsub.PubFactory.Threaded(100).Publish("/aggregate", pubsub.PubMultiWrap())
	go agg.Start()

	for {
		// WithInfo to detect a close.
		v, open := sub.ReadWithInfo()
		if !open {
			fmt.Println("\tWorker closing....")
			_ = agg.Close()
			return // All done
		}

		vi := v.(int64)
		if IsPrime(vi) {
			// Write all primes to the aggregator
			agg.Write(vi)
		}
	}
}

func Count() int64 {
	// Add a context so we can externally bind to the Done().
	// This is so we know when to read the value and exit.
	ctx, cancel := context.WithCancel(context.Background())

	// We only care about the count, the count subscriber just tracks the number
	// of items written to it.
	sub := pubsub.SubFactory.Counter().Subscribe("/aggregate", pubsub.SubContextWrap(cancel))

	// Wait for the subscriber to get called Done(), meaning all data is
	// published.
	<-ctx.Done()
	return sub.Count()
}

func IsPrime(i int64) bool {
	return big.NewInt(i).ProbablyPrime(0)
}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}
