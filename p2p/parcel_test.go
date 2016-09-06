package p2p_test

import (
	"testing"

	"bytes"
	"fmt"
	. "github.com/FactomProject/factomd/p2p"
)

func TestParcelMarshal(t *testing.T) {
	p := new(Parcel)
	var _ = p
	p.Header.Network = 1
	p.Header.Version = 2
	p.Header.Type = ParcelCommandType(3)
	p.Header.TargetPeer = "123"
	p.Header.NodeID = 4
	p.Header.PeerAddress = "456"
	p.Header.PeerPort = "789"
	p.Payload = []byte("This is a test")

	data, err := p.MarshalBinary()
	if err != nil {
		t.Fail()
	}

	fmt.Printf("Data: %x\n", data)

	p2 := new(Parcel)
	err = p2.UnmarshalBinary(data)
	if err != nil {
		t.Fail()
	}

	if len(data) != int(p2.Length) {
		t.Fail()
	}
	if p.Header.Network != p2.Header.Network {
		t.Fail()
	}
	if p.Header.Version != p2.Header.Version {
		t.Fail()
	}
	if p.Header.Type != p2.Header.Type {
		t.Fail()
	}
	if p.Header.NodeID != p2.Header.NodeID {
		t.Fail()
	}
	if p.Header.TargetPeer != p2.Header.TargetPeer {
		t.Fail()
	}
	if p.Header.PeerAddress != p2.Header.PeerAddress {
		t.Fail()
	}
	if p.Header.PeerPort != p2.Header.PeerPort {
		t.Fail()
	}
	if bytes.Compare(p.Payload, p2.Payload) != 0 {
		t.Fail()
	}

}
