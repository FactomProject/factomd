package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/FactomProject/factomd/pubsub/publishers"
	"github.com/FactomProject/factomd/pubsub/pubregistry"
	"github.com/FactomProject/factomd/pubsub/subscribers"
)

func main() {
	// This code finds the total number of prime numbers <= max
	max := int64(1e5)

	reg := pubregistry.NewRegistry()

	// One person publishes the work on a round robin basis
	robinPub := publishers.NewRoundRobinPublisher(100)
	go robinPub.Run() // The writer is threaded

	panicError(reg.Register("/source", robinPub))

	// Agg is a publisher with many writers.
	agg := publishers.NewSimpleMultiPublish(100)
	go agg.Run() // Writes are threaded
	panicError(reg.Register("/aggregate", agg))

	workers := 5
	for i := 0; i < workers; i++ {
		// Calculates if a number is prime. They will publish any primes
		// to the agg publisher. When the source publisher is done, and the
		// worker has no more tasks, it will close the agg publisher.
		// The agg publisher will close when all writers are done.
		go PrimeWorker(reg)
	}

	// Load the work
	go func() {
		for i := int64(0); i < max; i++ {
			robinPub.Write(i) // Write all numbers to the /source
		}
		_ = robinPub.Close() // Close when all numbers are done being read
	}()

	// Print the results. Count will wait for agg to be closed, then print the
	// count of things published to agg.
	fmt.Printf("%d primes found\n", Count(reg))
}

func PrimeWorker(reg *pubregistry.Registry) {
	// Channel based subscription is just a channel written to by a publisher
	sub := subscribers.NewChannelBasedSubscriber(5)
	panicError(reg.SubscribeTo("/source", sub))

	// agg is where we write our results.
	agg := reg.FindPublisher("/aggregate").(*publishers.SimpleMultiPublish)
	agg = agg.NewPublisher()

	for {
		// WithInfo to detect a close.
		v, open := sub.ReceiveWithInfo()
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

func Count(reg *pubregistry.Registry) int64 {
	// We only care about the count, the count subscriber just tracks the number
	// of items written to it.
	sub := subscribers.NewCounterSubscriber()

	// Add a context so we can externally bind to the Done().
	// This is so we know when to read the value and exit.
	ctx, cancel := context.WithCancel(context.Background())
	csub := subscribers.NewContext(sub, cancel)
	err := reg.SubscribeTo("/aggregate", csub)
	if err != nil {
		panic(err)
	}

	// Wait for the subscriber to get called Done(), meaning all data is
	// published.
	select {
	case <-ctx.Done():
		return sub.Count()
	}
}

func IsPrime(i int64) bool {
	return big.NewInt(i).ProbablyPrime(0)
}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}
