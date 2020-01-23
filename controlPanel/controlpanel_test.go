package controlpanel

import (
	"fmt"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/testHelper"
	"testing"
	"time"
)

// test for live testing
func TestControlPanelLive(t *testing.T) {
	// register the fnode so it can be retrieved
	s := testHelper.CreateEmptyTestState()
	fnode.New(s)

	// register the publisher to start the control panel
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

	New(s.FactomNodeName)

	select {}
}
