// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

//A placeholder structure for messages
type RevealEntryMsg struct {
	MessageBase
	Timestamp interfaces.Timestamp
	Entry     interfaces.IEntry

	//Not marshalled
	hash        interfaces.IHash
	chainIDHash interfaces.IHash
	isEntry     bool
	commitChain *CommitChainMsg
	commitEntry *CommitEntryMsg
}

var _ interfaces.IMsg = (*RevealEntryMsg)(nil)

func (m *RevealEntryMsg) Process(dbheight uint32, state interfaces.IState) bool {
	commit := state.GetCommits(m.GetHash())
	if commit == nil {
		panic("commit was nil in process, this should not happen")
	}

	if _, isNewChain := commit.(*CommitChainMsg); isNewChain {
		fmt.Println("New Chain")
		chainID := m.Entry.GetChainID()
		eb, err := state.GetDB().FetchEBlockHead(chainID)
		if err != nil || eb != nil {
			panic("Chain already exists")
		}

		// Create a new Entry Block for a new Entry Block Chain
		eb = entryBlock.NewEBlock()
		// Set the Chain ID
		eb.GetHeader().SetChainID(m.Entry.GetChainID())
		// Set the Directory Block Height for this Entry Block
		eb.GetHeader().SetDBHeight(dbheight)
		// Add our new entry
		eb.AddEBEntry(m.Entry)
		// Put it in our list of new Entry Blocks for this Directory Block
		state.PutNewEBlocks(dbheight, m.Entry.GetChainID(), eb)
		state.PutNewEntries(dbheight, m.Entry.GetHash(), m.Entry)
		
		return true
	} else if _, isNewEntry := commit.(*CommitEntryMsg); isNewEntry {
		chainID := m.Entry.GetChainID()
		eb := state.GetNewEBlocks(dbheight, chainID)
		if eb == nil {
			prev, err := state.GetDB().FetchEBlockHead(chainID)
			if prev == nil || err != nil {
				return false
			}
			eb = entryBlock.NewEBlock()
			// Set the Chain ID
			eb.GetHeader().SetChainID(m.Entry.GetChainID())
			// Set the Directory Block Height for this Entry Block
			eb.GetHeader().SetDBHeight(dbheight)
			// Set the PrevKeyMR
			key, _ := prev.KeyMR()
			eb.GetHeader().SetPrevKeyMR(key)
		}
		// Add our new entry
		eb.AddEBEntry(m.Entry)
		// Put it in our list of new Entry Blocks for this Directory Block
		state.PutNewEBlocks(dbheight, m.Entry.GetChainID(), eb)
		state.PutNewEntries(dbheight, m.Entry.GetHash(), m.Entry)

		return true
	}
	log.Println("Found Bad Commit")
	return false
}

func (m *RevealEntryMsg) GetHash() interfaces.IHash {
	if m.hash == nil {
		m.hash = m.Entry.GetHash()
	}
	return m.hash
}

func (m *RevealEntryMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *RevealEntryMsg) GetChainIDHash() interfaces.IHash {
	if m.chainIDHash == nil {
		m.chainIDHash = primitives.Sha(m.Entry.GetChainID().Bytes())
	}
	return m.chainIDHash
}

func (m *RevealEntryMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *RevealEntryMsg) Type() int {
	return constants.REVEAL_ENTRY_MSG
}

func (m *RevealEntryMsg) Int() int {
	return -1
}

func (m *RevealEntryMsg) Bytes() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *RevealEntryMsg) Validate(state interfaces.IState) int {
	commit := state.GetCommits(m.GetHash())
	ECs := 0

	if commit == nil {
		return 0
	}

	//
	// Make sure one of the two proper commits got us here.
	var okChain, okEntry bool
	m.commitChain, okChain = commit.(*CommitChainMsg)
	m.commitEntry, okEntry = commit.(*CommitEntryMsg)
	if !okChain && !okEntry {
		return -1
	}

	// Now make sure the proper amount of credits were paid to record the entry.
	if okEntry {
		m.isEntry = true
		ECs = int(m.commitEntry.CommitEntry.Credits)
		if m.Entry.KSize() < ECs {
			fmt.Println("KSize", m.Entry.KSize(), ECs)
			return -1
		}
	} else {
		m.isEntry = false
		ECs = int(m.commitChain.CommitChain.Credits)
		if m.Entry.KSize()+10 < ECs {
			fmt.Println("KSize", m.Entry.KSize(), ECs)
			return -1
		}
	}

	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *RevealEntryMsg) Leader(state interfaces.IState) bool {
	return state.LeaderFor(m.GetHash().Bytes())
}

// Execute the leader functions of the given message
func (m *RevealEntryMsg) LeaderExecute(state interfaces.IState) error {
	c := state.GetCommits(m.GetHash())
	if c != nil {
		return state.LeaderExecute(m)
	}
	state.PutReveals(m.GetHash(), m)

	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *RevealEntryMsg) Follower(interfaces.IState) bool {
	return true
}

func (m *RevealEntryMsg) FollowerExecute(state interfaces.IState) error {
	matched, err := state.FollowerExecuteMsg(m)
	if err != nil {
		return err
	}
	if matched { // We matched, we must be remembered!
		fmt.Println("Matched!")
	} else {
		fmt.Println("Not Matched!")
	}
	return nil
}

func (e *RevealEntryMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RevealEntryMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *RevealEntryMsg) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func NewRevealEntryMsg() *RevealEntryMsg {
	return new(RevealEntryMsg)
}

func (m *RevealEntryMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data[1:]
	e := entryBlock.NewEntry()
	newData, err = e.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.Entry = e
	return newData, nil
}

func (m *RevealEntryMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *RevealEntryMsg) MarshalBinary() (data []byte, err error) {
	data, err = m.Entry.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data = append([]byte{byte(m.Type())}, data...)
	return data, nil
}

func (m *RevealEntryMsg) String() string {
	return "RevealEntryMsg " + m.Timestamp.String() + " " + m.GetHash().String()
}
