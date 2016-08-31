// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// factomlog is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package anchor

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
)

//Construct the entry and submit it to the server
func (a *Anchor) submitEntryToAnchorChain(aRecord *AnchorRecord) error {
	jsonARecord, err := json.Marshal(aRecord)
	//anchorLog.Debug("submitEntryToAnchorChain - jsonARecord: ", string(jsonARecord))
	if err != nil {
		return err
	}
	bufARecord := new(primitives.Buffer)
	bufARecord.Write(jsonARecord)
	aRecordSig := a.serverPrivKey.Sign(jsonARecord)

	//Create a new entry
	entry := entryBlock.NewEntry()
	entry.ChainID = a.anchorChainID
	anchorLog.Debug("anchorChainID: ", a.anchorChainID)
	// instead of append signature at the end of anchor record
	// it can be added as the first entry.ExtIDs[0]
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: aRecordSig.Bytes()})
	entry.Content = primitives.ByteSlice{Bytes: bufARecord.DeepCopyBytes()}
	//anchorLog.Debug("entry: ", spew.Sdump(entry))

	buf := new(primitives.Buffer)
	// 1 byte version
	buf.Write([]byte{0})
	// 6 byte milliTimestamp (truncated unix time)
	buf.Write(milliTime())
	// 32 byte Entry Hash
	buf.Write(entry.GetHash().Bytes())
	// 1 byte number of entry credits to pay
	binaryEntry, err := entry.MarshalBinary()
	if err != nil {
		return err
	}

	anchorLog.Debug("jsonARecord binary entry: ", hex.EncodeToString(binaryEntry))
	if c, err := util.EntryCost(binaryEntry); err == nil {
		buf.WriteByte(byte(c))
	} else {
		return err
	}

	tmp := buf.DeepCopyBytes()
	sig := a.serverECKey.Sign(tmp).(*primitives.Signature)
	buf = primitives.NewBuffer(tmp)
	buf.Write(a.serverECKey.Pub[:])
	buf.Write(sig.Sig[:])

	commit := entryCreditBlock.NewCommitEntry()
	err = commit.UnmarshalBinary(buf.DeepCopyBytes())
	if err != nil {
		return err
	}

	// create a CommitEntry msg and send it to the local inmsgQ
	cm := messages.NewCommitEntryMsg()
	cm.CommitEntry = commit
	a.state.InMsgQueue() <- cm

	// create a RevealEntry msg and send it to the local inmsgQ
	rm := messages.NewRevealEntryMsg()
	rm.Entry = entry
	a.state.InMsgQueue() <- rm

	return nil
}

// MilliTime returns a 6 byte slice representing the unix time in milliseconds
func milliTime() (r []byte) {
	buf := new(primitives.Buffer)
	t := time.Now().UnixNano()
	m := t / 1e6
	binary.Write(buf, binary.BigEndian, m)
	return buf.DeepCopyBytes()[2:]
}
