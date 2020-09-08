package main

import (
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/modules/pubsub"
)

// Shows an example for maintaining a config
func main() {
	initialConfig := map[string]interface{}{
		"blktime": 600,
		"network": "LOCAL",
	}

	// establish a publisher
	configPublisher := pubsub.PubFactory.Base().Publish("/source", pubsub.PubInitMapWrap(initialConfig), pubsub.PubMultiWrap())

	// Just to test the multi wrap
	anotherConfigPublisher := pubsub.PubFactory.Base().Publish("/source", pubsub.PubInitMapWrap(initialConfig), pubsub.PubMultiWrap())

	// establish a subscriber to read the value
	subscriber := pubsub.SubFactory.Config(func(o map[string]interface{}) {
		// A print for each new value
		fmt.Printf("New update! %v\n", o)
	}).Subscribe("/source")

	// First check the initials
	blktime, _ := subscriber.Int("blktime")
	if blktime != initialConfig["blktime"].(int) {
		panic("Unexpected initial blktime")
	}

	network, _ := subscriber.String("network")
	if network != initialConfig["network"].(string) {
		panic("Unexpected initial network")
	}

	// Send some updates
	configPublisher.Write(map[string]interface{}{"random": 1234})
	random, _ := subscriber.Int("random")
	fmt.Println("Found -> random=", random) // 1234

	anotherConfigPublisher.Write(map[string]interface{}{"blktime": 60})
	blktime, _ = subscriber.Int("blktime")
	fmt.Println("Found -> blktime=", blktime)

	// Send some we already know
	anotherConfigPublisher.Write(map[string]interface{}{"blktime": 60})
	anotherConfigPublisher.Write(map[string]interface{}{"random": 1234})

	// A new subscriber should get the latest
	subscriber2 := pubsub.SubFactory.Config(func(o map[string]interface{}) {
		// A print for each new value
		fmt.Printf("[2] New update! %v\n", o)
	}).Subscribe("/source")
	if !reflect.DeepEqual(subscriber.Read(), subscriber2.Read()) {
		panic("Configs differ")
	}
	fmt.Printf("New Sub: %v\n", subscriber2.Read())

}
