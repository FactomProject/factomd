package controlpanel

import (
	"fmt"
	"github.com/FactomProject/factomd/pubsub"
	"testing"
	"time"
)

func TestControlPanel(t *testing.T) {
	pubsub.ResetGlobalRegistry()
	p := pubsub.PubFactory.Threaded(5).Publish("test")
	go p.Start()

	go func() {
		i := 1
		for {
			p.Write(fmt.Sprintf("data: %d", i))
			time.Sleep(2 * time.Second)
			i++
		}
	}()

	ControlPanel()

	select {}
}
