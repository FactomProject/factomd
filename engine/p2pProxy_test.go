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
