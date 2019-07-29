package eventservices

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventmessages"
)

func mapCommitChain(msg interfaces.IMsg) *eventmessages.FactomEvent_CommitChain {
	commitChain := msg.(*messages.CommitChainMsg).CommitChain
	ecPubKey := commitChain.ECPubKey.Fixed()
	sig := commitChain.Sig

	result := &eventmessages.FactomEvent_CommitChain{
		CommitChain: &eventmessages.CommitChain{
			ChainIDHash: &eventmessages.Hash{
				HashValue: commitChain.ChainIDHash.Bytes()},
			EntryHash: &eventmessages.Hash{
				HashValue: commitChain.EntryHash.Bytes(),
			},
			Timestamp: convertByteSlice6ToTimestamp(commitChain.MilliTime),
			Credits:   uint32(commitChain.Credits),
			EcPubKey:  ecPubKey[:],
			Sig:       sig[:],
		}}
	return result
}
