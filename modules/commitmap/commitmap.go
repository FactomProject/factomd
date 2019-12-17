package commitmap

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/pubsubtypes"
	"github.com/FactomProject/factomd/generated"
	"github.com/FactomProject/factomd/pubsub"
)

type commitStatus struct {
	iMsg     interfaces.IMsg
	revealed bool
}

type CommitMap struct {
	common.Name
	commits map[[32]byte]commitStatus

	// New DependentHolding
	addMsg          *generated.Subscribe_ByChannel_CommitRequest_type // commit/reveal messages from VMs to be added to map
	checkHash       *generated.Subscribe_ByChannel_CommitRequest_type // check for commit to exits
	leaderTimestamp *generated.Subscribe_ByValue_Timestamp_type       // Current Leader Timestamp
}

func NewCommitMap(parent common.NamedObject, instance int) *CommitMap {
	b := new(CommitMap)
	b.NameInit(parent, fmt.Sprintf("commitmap%d", instance), reflect.TypeOf(b).String())
	b.addMsg = generated.Subscribe_ByChannel_CommitRequest(pubsub.SubFactory.Channel(100))    //.Subscribe("path?")
	b.checkHash = generated.Subscribe_ByChannel_CommitRequest(pubsub.SubFactory.Channel(100)) //.Subscribe("path?")
	b.leaderTimestamp = generated.Subscribe_ByValue_Timestamp(pubsub.SubFactory.Value())
	// All dependent holdings in an fnode publish into one multiwrap
	//path := pubsub.GetPath(b.Name.GetParentName(), "commitmap", "msgout")
	//b.outMsgs = generated.Publish_PubBase_IMsg(pubsub.PubFactory.Threaded(100).Publish(path, pubsub.PubMultiWrap()))
	return b
}

func (b *CommitMap) Publish() {
	//	go b.outMsgs.Start()
}
func (b *CommitMap) Subscribe() {
	// TODO: Find actual paths
	b.addMsg.SubChannel.Subscribe(pubsub.GetPath(b.GetParentName(), "commits"))
	b.checkHash.SubChannel.Subscribe(pubsub.GetPath(b.GetParentName(), "checkcommits"))
}

func (b *CommitMap) ClosePublishing() {
	//_ = b.outMsgs.Close()
}

func (b *CommitMap) get(hash [32]byte) (valid bool, revealed bool, iMsg interfaces.IMsg) {
	status, ok := b.commits[hash]
	// if there is a commit check if it is outdated.
	if ok {
		leaderTime := b.leaderTimestamp.Read()
		if status.iMsg.GetTimestamp().GetTimeMilli() < leaderTime.GetTimeMilli()-60*60*1000 {
			delete(b.commits, hash)
			return false, false, nil
		}
	}
	return ok, status.revealed, status.iMsg
}

// Handle checking a CommitEntry or CommitChain or Reveal to see if its valid to add to the Commit Map
func (b *CommitMap) handleHash(iMsg interfaces.IMsg) error {
	hash := iMsg.GetHash().Fixed()
	ok, revealed, commit := b.get(hash)
	switch iMsg.Type() {
	case constants.REVEAL_ENTRY_MSG:
		if !ok {
			return errors.New("reveal before commit")
		}
		if revealed {
			return errors.New("reveal already revealed") // toss the new commit if it hits an unexpired revealed commit
		}
		return nil // Good to go for reveal.

	case constants.COMMIT_ENTRY_MSG, constants.COMMIT_CHAIN_MSG:
		if revealed {
			return errors.New("commit already revealed") // toss the new commit if it hits an unexpired revealed commit
		}
		fee0, fee1 := getFees(iMsg, commit)
		if fee0 > fee1 { // new commit is higher paid so keep it instead
			return errors.New("duplicate lower fee") // toss the new commit if it hits an unexpired revealed commit
		}
	}
}

func getFees(iMsg interfaces.IMsg, commit interfaces.IMsg) (uint8, uint8) {
	var fee0, fee1 uint8
	switch iMsg.Type() {
	case constants.COMMIT_ENTRY_MSG:
		fee0 = iMsg.(*messages.CommitEntryMsg).CommitEntry.Credits
		fee1 = commit.(*messages.CommitEntryMsg).CommitEntry.Credits

	case constants.COMMIT_CHAIN_MSG:
		fee0 = iMsg.(*messages.CommitChainMsg).CommitChain.Credits
		fee1 = commit.(*messages.CommitChainMsg).CommitChain.Credits
	}
	return fee0, fee1
}

// Handle a CommitEntry or CommitChain or Reveal being added to the commit map
func (b *CommitMap) handleIMsg(iMsg interfaces.IMsg) error {
	hash := iMsg.GetHash().Fixed()
	ok, revealed, commit := b.get(hash)
	switch iMsg.Type() {
	case constants.REVEAL_ENTRY_MSG:
		if !ok {
			panic("reveal before commit")
		}
		if revealed {
			panic("already revealed")
		}
		b.commits[hash].revealed = true
		return nil // added to commit map

	case constants.COMMIT_ENTRY_MSG, constants.COMMIT_CHAIN_MSG:
		if !ok {
			b.commits[hash] = commitStatus{iMsg: iMsg, revealed: false}
			return nil
		}
		if revealed {
			return errors.New("Commit Already Revealed") // toss the new commit if it hits an unexpired revealed commit
		}
		if iMsg.Type() != commit.Type() {
			panic("mismatched commit types")
		}
		fee0, fee1 := getFees(iMsg, commit)
		if fee0 > fee1 { // new commit is higher paid so keep it instead
			return errors.New("duplicate lower fee") // toss the new commit if it hits an unexpired revealed commit
		}
		b.commits[hash] = commitStatus{iMsg: iMsg, revealed: false}
		return nil // Added to the commit map
	default:
		panic("unexpected type")
	}
}

func (b *CommitMap) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-b.addMsg.Channel():
			req := data.(pubsubtypes.CommitRequest)
			req.Channel <- b.handleIMsg(req.IMsg)
		case data := <-b.checkHash.Channel():
			req := data.(pubsubtypes.CommitRequest)
			req.Channel <- b.handleCheck(req.Hash)
		}

	}
}
