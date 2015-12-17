// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// factomlog is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package anchor

import (
	"bytes"
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
func submitEntryToAnchorChain(aRecord *AnchorRecord) error {
	jsonARecord, err := json.Marshal(aRecord)
	//anchorLog.Debug("submitEntryToAnchorChain - jsonARecord: ", string(jsonARecord))
	if err != nil {
		return err
	}
	bufARecord := new(bytes.Buffer)
	bufARecord.Write(jsonARecord)
	aRecordSig := serverPrivKey.Sign(jsonARecord)

	//Create a new entry
	entry := entryBlock.NewEntry()
	entry.ChainID = anchorChainID
	anchorLog.Debug("anchorChainID: ", anchorChainID)
	// instead of append signature at the end of anchor record
	// it can be added as the first entry.ExtIDs[0]
	entry.ExtIDs = append(entry.ExtIDs, aRecordSig.Bytes())
	entry.Content = bufARecord.Bytes()
	//anchorLog.Debug("entry: ", spew.Sdump(entry))

	buf := new(bytes.Buffer)
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

	tmp := buf.Bytes()
	sig := serverECKey.Sign(tmp).(*primitives.Signature)
	buf = bytes.NewBuffer(tmp)
	buf.Write(serverECKey.Pub.Key[:])
	buf.Write(sig.Sig[:])

	commit := entryCreditBlock.NewCommitEntry()
	err = commit.UnmarshalBinary(buf.Bytes())
	if err != nil {
		return err
	}

	// create a CommitEntry msg and send it to the local inmsgQ
	cm := messages.NewCommitEntryMsg()
	cm.CommitEntry = commit
	state.NetworkInMsgQueue() <- cm

	// create a RevealEntry msg and send it to the local inmsgQ
	rm := messages.NewRevealEntryMsg()
	rm.Entry = entry
	state.NetworkInMsgQueue() <- rm

	return nil
}

// MilliTime returns a 6 byte slice representing the unix time in milliseconds
func milliTime() (r []byte) {
	buf := new(bytes.Buffer)
	t := time.Now().UnixNano()
	m := t / 1e6
	binary.Write(buf, binary.BigEndian, m)
	return buf.Bytes()[2:]
}
