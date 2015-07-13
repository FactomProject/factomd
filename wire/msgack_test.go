package wire_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/wallet"
	"github.com/FactomProject/btcd/wire"
	"testing"
	"time"
)

func TestCommitChain(t *testing.T) {
	fmt.Println("\nTestAck===========================================================================")
	ack := wire.NewMsgAcknowledgement(1, 2, nil, END_MINUTE_10) //??

	// Sign the ack using server private keys
	bytes, _ := ack.GetBinaryForSignature()
	ack.Signature = *plMgr.SignAck(bytes).Sig
	//??

	if err != nil {
		t.Errorf("Error:", err)
	}

}
