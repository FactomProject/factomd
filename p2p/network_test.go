package p2p

import (
	"fmt"
	"testing"
)

var unitTestNetworks = 1

func testNetworkHarness(t *testing.T) *Network {
	conf := DefaultP2PConfiguration()
	conf.SeedURL = ""
	conf.NodeName = fmt.Sprintf("UnitTestNode-%d", unitTestNetworks)
	conf.Network = NewNetworkID("unit-test-network")
	conf.Special = ""
	conf.EnablePrometheus = false
	unitTestNetworks++

	n, err := NewNetwork(conf)
	if err != nil {
		t.Fatal(err)
	}
	return n
}
