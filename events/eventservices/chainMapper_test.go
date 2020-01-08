package eventservices

import (
	"testing"

	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/stretchr/testify/assert"
)

func TestMapCommitChain(t *testing.T) {
	msg := newTestCommitChainMsg()

	factomCommitChain := mapCommitChain(eventmessages.EntityState_ACCEPTED, msg)

	assert.NotNil(t, factomCommitChain)
	assert.NotNil(t, factomCommitChain.ChainCommit)
	assert.NotNil(t, factomCommitChain.ChainCommit.Version)
	assert.NotNil(t, factomCommitChain.ChainCommit.Timestamp)
	assert.NotNil(t, factomCommitChain.ChainCommit.EntryHash)
	assert.NotNil(t, factomCommitChain.ChainCommit.ChainIDHash)
	assert.NotNil(t, factomCommitChain.ChainCommit.EntryCreditPublicKey)
	assert.NotNil(t, factomCommitChain.ChainCommit.Signature)
	assert.NotNil(t, factomCommitChain.ChainCommit.Credits)
	assert.NotNil(t, factomCommitChain.ChainCommit.Weld)

	assert.Equal(t, eventmessages.EntityState_ACCEPTED, factomCommitChain.ChainCommit.EntityState)
}

func TestMapCommitChainState(t *testing.T) {
	msg := newTestCommitChainMsg()

	factomStateChange := mapCommitChainState(eventmessages.EntityState_REQUESTED, msg)

	assert.NotNil(t, factomStateChange)
	assert.NotNil(t, factomStateChange.StateChange)
	assert.NotNil(t, factomStateChange.StateChange.EntityState)
	assert.NotNil(t, factomStateChange.StateChange.BlockHeight)
	assert.NotNil(t, factomStateChange.StateChange.EntityHash)
}

func newTestCommitChainMsg() *messages.CommitChainMsg {
	msg := new(messages.CommitChainMsg)
	msg.Signature = nil
	msg.CommitChain = entryCreditBlock.NewCommitChain()
	msg.CommitChain.ChainIDHash.SetBytes([]byte(""))
	msg.CommitChain.ECPubKey = new(primitives.ByteSlice32)
	msg.CommitChain.Sig = new(primitives.ByteSlice64)
	msg.CommitChain.Weld.SetBytes([]byte("1"))
	return msg
}
