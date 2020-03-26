package main

import (
	"fmt"

	"github.com/FactomProject/factomd/modules/pubsub"
)

// Absolute basic minimum
func main() {
	// establish a publisher
	//                             Type             Path
	publisher1 := pubsub.PubFactory.Base().Publish("/source", pubsub.PubMultiWrap())
	publisher2 := pubsub.PubFactory.Base().Publish("/source", pubsub.PubMultiWrap())

	// establish a subscriber to read the value
	subscriber := pubsub.SubFactory.Value().Subscribe("/source")

	publisher1.Write(1)
	fmt.Println(subscriber.Read()) // 1

	publisher2.Write(2)
	fmt.Println(subscriber.Read()) // 1
}
