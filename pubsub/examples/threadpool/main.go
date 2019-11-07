package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/FactomProject/factomd/pubsub"
)

func main() {
	// This code finds the total number of prime numbers <= max
	max := int64(1e5)

	reg := pubsub.NewRegistry()

	// One person publishes the work on a round robin basis
	robinPub := pubsub.NewRoundRobinPublisher(100).Publish("/source")
	go robinPub.Run() // The writer is threaded

	// Agg is a publisher with many writers.
	agg := pubsub.NewSimpleMultiPublish(100).Publish("/aggregate")
	go agg.Run() // Writes are threaded

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
		for i := int64(0); i < max; i++ {
			robinPub.Write(i) // Write all numbers to the /source
		}
		_ = robinPub.Close() // Close when all numbers are done being read
	}()

	// Print the results. Count will wait for agg to be closed, then print the
	// count of things published to agg.
	fmt.Printf("%d primes found\n", Count(reg))
}

func PrimeWorker() {
	// Channel based subscription is just a channel written to by a publisher
	sub := pubsub.NewChannelBasedSubscriber(5).Subscribe("/source")

	// agg is where we write our results.
	agg := pubsub.GlobalRegistry().FindPublisher("/aggregate").(*pubsub.PubSimpleMulti)
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

func Count(reg *pubsub.Registry) int64 {
	// Add a context so we can externally bind to the Done().
	// This is so we know when to read the value and exit.
	ctx, cancel := context.WithCancel(context.Background())

	// We only care about the count, the count subscriber just tracks the number
	// of items written to it.
	sub := pubsub.NewCounterSubscriber().Subscribe("/aggregate", pubsub.NewContextWrap(cancel))

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
