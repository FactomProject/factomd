package wire_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/btcd/wire"
	"testing"
)

func TestCommitEntry(t *testing.T) {

	//Write msg
	msgOutgoing := wire.NewMsgCommitEntry()
	msgOutgoing.Version = 0
	if p, err := hex.DecodeString("13dc481ea11f"); err != nil {
		t.Error(err)
	} else {
		copy(msgOutgoing.MilliTime[:], p[:6])
	}
	if p, err := hex.DecodeString("1af341223423eb25edc394ceeb6c5fb7f42f9f22640bc29b7a2e949f5dc68563"); err != nil {
		t.Error(err)
	} else {
		msgOutgoing.EntryHash.SetBytes(p[:32])
	}
	msgOutgoing.Credits = 1
	if p, err := hex.DecodeString("bf11aac8394113ecfb6290d9eba0d6995b34fab963cb1d9c6b30d6f850111875"); err != nil {
		t.Error(err)
	} else {
		copy(msgOutgoing.ECPubKey[:], p[:32])
	}
	if p, err := hex.DecodeString("36051a49807fa903a13747223ee67205e6d75868923daa99182788adaf7fa7eec6112af56772833a76f01caf578d516f7c938d2c4b655f6d4a7666542f32000b"); err != nil {
		t.Error(err)
	} else {
		copy(msgOutgoing.Sig[:], p[:64])
	}
	fmt.Printf("msgOutgoing:%+v\n", msgOutgoing)

	var buf bytes.Buffer
	msgOutgoing.BtcEncode(&buf, uint32(1))
	fmt.Printf("Outgoing msg bytes: %x\n", buf.Bytes())

	//Read msg
	msgIncoming := wire.NewMsgCommitEntry()
	err := msgIncoming.BtcDecode(&buf, uint32(1))
	fmt.Printf("msgIncoming:%+v\n", msgIncoming)
	m := msgIncoming
	fmt.Println(m.Version)
	fmt.Printf("%x\n", m.MilliTime)
	fmt.Printf("%x\n", m.EntryHash.Bytes)
	fmt.Println(m.Credits)
	fmt.Printf("%x\n", m.ECPubKey)
	fmt.Printf("%x\n", m.Sig)

	if err != nil {
		t.Errorf("Error:", err)
	}
}
