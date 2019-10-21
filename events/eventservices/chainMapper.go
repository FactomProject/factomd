package eventservices

import (
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
)

func mapCommitChain(entityState eventmessages.EntityState, commitChainMsg *messages.CommitChainMsg) *eventmessages.FactomEvent_ChainCommit {
	commitChain := commitChainMsg.CommitChain
	ecPubKey := commitChain.ECPubKey.Fixed()
	sig := commitChain.Sig

	result := &eventmessages.FactomEvent_ChainCommit{
		ChainCommit: &eventmessages.ChainCommit{
			EntityState:          entityState,
			ChainIDHash:          commitChain.ChainIDHash.Bytes(),
			EntryHash:            commitChain.EntryHash.Bytes(),
			Timestamp:            convertByteSlice6ToTimestamp(commitChain.MilliTime),
			Credits:              uint32(commitChain.Credits),
			EntryCreditPublicKey: ecPubKey[:],
			Signature:            sig[:],
			Version:              uint32(commitChain.Version),
			Weld:                 commitChain.Weld.Bytes(),
		}}
	return result
}

func mapCommitChainState(state eventmessages.EntityState, commitChainMsg *messages.CommitChainMsg) *eventmessages.FactomEvent_StateChange {
	result := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash:  commitChainMsg.CommitChain.ChainIDHash.Bytes(),
			EntityState: state,
		},
	}
	return result
}
