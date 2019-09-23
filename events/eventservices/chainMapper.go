package eventservices

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
)

func mapCommitChain(entityState eventmessages.EntityState, msg interfaces.IMsg) *eventmessages.FactomEvent_ChainCommit {
	commitChain := msg.(*messages.CommitChainMsg).CommitChain
	ecPubKey := commitChain.ECPubKey.Fixed()
	sig := commitChain.Sig

	result := &eventmessages.FactomEvent_ChainCommit{
		ChainCommit: &eventmessages.ChainCommit{
			EntityState: entityState,
			ChainIDHash: &eventmessages.Hash{
				HashValue: commitChain.ChainIDHash.Bytes()},
			EntryHash: &eventmessages.Hash{
				HashValue: commitChain.EntryHash.Bytes(),
			},
			Timestamp:            convertByteSlice6ToTimestamp(commitChain.MilliTime),
			Credits:              uint32(commitChain.Credits),
			EntryCreditPublicKey: ecPubKey[:],
			Signature:            sig[:],
			Version:              uint32(commitChain.Version),
		}}
	return result
}

func mapCommitChainState(state eventmessages.EntityState, msg interfaces.IMsg) *eventmessages.FactomEvent_StateChange {
	commitChain := msg.(*messages.CommitChainMsg).CommitChain
	result := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash: &eventmessages.Hash{
				HashValue: commitChain.ChainIDHash.Bytes()},
			EntityState: state,
		},
	}
	return result
}
