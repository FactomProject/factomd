package engine_test

import (
	"testing"

	. "github.com/FactomProject/factomd/engine"
)

func TestFactomMessage(t *testing.T) {
	testMessage := FactomMessage{
		Message:  []byte("msgtest"),
		PeerHash: "peertest",
		AppHash:  "hashtest",
		AppType:  "typetest",
	}

	_ = testMessage.String()

	_, err := testMessage.JSONByte()
	if err != nil {
		t.Error(err)
	}
}

func TestMiscP2Pproxy(t *testing.T) {
	p2pTestProxy := new(P2PProxy).Initialize("FNodeTest", "P2P Network").(*P2PProxy)

	if p2pTestProxy.Weight() != 0 {
		t.Error("Weight should be 0 on newly-initialized p2pProxy instance")
	}

	if p2pTestProxy.BytesOut() != 0 {
		t.Error("BytesOut should be 0 on newly-initialized p2pProxy instance")
	}

	if p2pTestProxy.BytesIn() != 0 {
		t.Error("BytesIn should be 0 on newly-initialized p2pProxy instance")
	}

	if p2pTestProxy.Len() != 0 {
		t.Error("Len should be 0 on newly-initialized p2pProxy instance")
	}

	p2pTestProxy.SetWeight(2)
	if p2pTestProxy.Weight() != 2 {
		t.Error("Weight should be 2 on after calling p2pProxy.SetWeight(2)")
	}

	nameFrom := p2pTestProxy.GetNameFrom()
	if nameFrom != "FNodeTest" {
		t.Errorf("GetNameFrom was %s instead of FNodeTest\n", nameFrom)
	}

}
