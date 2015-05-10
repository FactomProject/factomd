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

func TestCommitEntry(t *testing.T) {
	fmt.Println("\nTestCommitEntry===========================================================================")
	bName := make([][]byte, 0, 5)
	bName = append(bName, []byte("myCompany"))
	bName = append(bName, []byte("bookkeeping2"))

	chainID,_ := common.GetChainID(bName)

	entry := new(common.Entry)
	entry.ExtIDs = bName
	entry.Data = []byte("Entry data: asl;djfasldkfjasldfjlksouiewopurw222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222\"")
	entry.ChainID = chainID

	binaryEntry, _ := entry.MarshalBinary()
	entryHash := common.Sha(binaryEntry)

	// Calculate the required credits
	credits := uint32(binary.Size(binaryEntry)/1000 + 1)

	timestamp := uint64(time.Now().Unix())
	var msg bytes.Buffer
	binary.Write(&msg, binary.BigEndian, timestamp)
	msg.Write(entryHash.Bytes)
	binary.Write(&msg, binary.BigEndian, credits)

	sig := wallet.SignData(msg.Bytes())

	hexkey := "ed14447c656241bf7727fce2e2a48108374bec6e71358f0a280608b292c7f3bc"
	binkey, _ := hex.DecodeString(hexkey)
	pubKey := new(common.Hash)
	pubKey.SetBytes(binkey)

	//Write msg
	msgOutgoing := wire.NewMsgCommitEntry()
	msgOutgoing.ECPubKey = pubKey
	msgOutgoing.EntryHash = entryHash
	msgOutgoing.Credits = credits
	msgOutgoing.Timestamp = timestamp
	msgOutgoing.Sig = sig.Sig[:]
	fmt.Printf("msgOutgoing:%+v\n", msgOutgoing)

	var buf bytes.Buffer
	msgOutgoing.BtcEncode(&buf, uint32(1))
	fmt.Println("Outgoing msg bytes: ", buf.Bytes())

	//Read msg
	msgIncoming := wire.NewMsgCommitEntry()
	err := msgIncoming.BtcDecode(&buf, uint32(1))

	fmt.Printf("msgIncoming:%+v\n", msgIncoming)

	if err != nil {
		t.Errorf("Error:", err)
	}

}
