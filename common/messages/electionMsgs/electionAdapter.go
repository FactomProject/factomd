package electionMsgs

import (
	"crypto/sha256"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	primitives2 "github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/electionsCore/election"
	"github.com/FactomProject/factomd/electionsCore/imessage"
	"github.com/FactomProject/factomd/electionsCore/messages"
	"github.com/FactomProject/factomd/electionsCore/primitives"
	llog "github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
)

// ElectionAdapter is used to drive the election package, abstracting away factomd
// logic and messages
type ElectionAdapter struct {
	Election *elections.Elections

	Electing int
	DBHeight int
	Minute   int
	VMIndex  int

	// Processed indicates the election completed and was processed
	// AKA leader was swapper
	ElectionProcessed bool // On Election
	StateProcessed    bool // On State

	// All messages we adapt so we can expand them
	taggedMessages map[[32]byte]interfaces.IMsg

	// We need these to expand our own votes
	Volunteers map[[32]byte]*FedVoteVolunteerMsg

	// Audits kept as IHash for easy access (cache)
	AuditServerList   []interfaces.IHash
	SimulatedElection *election.Election
}

func (ea *ElectionAdapter) VolunteerControlsStatus() (status string) {
	defer func() {
		if r := recover(); r != nil {
			llog.LogPrintf("recovery", "Fail in VolunteerControlsStatus %v", r)
			status = ""
			return
		}
	}()
	return ea.SimulatedElection.VolunteerControlString()
}

func (ea *ElectionAdapter) MessageLists() string {
	return fmt.Sprintf("Election-  DBHeight: %d, Minut: %d, Messages %d\n%s", ea.DBHeight, ea.Minute, ea.Electing, ea.SimulatedElection.PrintMessages())
}

func (ea *ElectionAdapter) Status() string {
	return fmt.Sprintf("Election-  DBHeight: %d, Minut: %d, Electing %d\n%s", ea.DBHeight, ea.Minute, ea.Electing, ea.SimulatedElection.Display.String())
}

func (ea *ElectionAdapter) GetAudits() []interfaces.IHash {
	return ea.AuditServerList
}

// Compare two Ids
func lessId(a, b primitives.Identity) bool {
	for i, x := range b {
		if a[i] != x {
			return a[i] < x // first unequal byte determines order
		}
	}
	return false
}

// Xor a mask with an ID
func maskId(mask, b primitives.Identity) primitives.Identity {
	for i, x := range b {
		b[i] = x ^ mask[i]
	}
	return b
}

func buildPriorityOrder(audits []interfaces.IServer, dbHash interfaces.IHash, minute int, vm int) (arr []primitives.Identity) {
	// find a randomizing mask
	mask := sha256.Sum256(append(dbHash.Bytes(), byte(minute), byte(vm)))

	// build a list with the mask applied
	for _, a := range audits {
		arr = append(arr, maskId(mask, a.GetChainID().Fixed()))
	}

	// bubble sort with the mask
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(arr)-i; j++ {
			if lessId(arr[j], arr[j+1]) {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}

	// remove the mask
	for i, a := range arr {
		arr[i] = maskId(mask, a)
	}
	return arr
}

func NewElectionAdapter(e *elections.Elections, dbHash interfaces.IHash) *ElectionAdapter {
	ea := new(ElectionAdapter)
	ea.taggedMessages = make(map[[32]byte]interfaces.IMsg)
	ea.Volunteers = make(map[[32]byte]*FedVoteVolunteerMsg)

	ea.DBHeight = e.DBHeight
	ea.Minute = e.Minute
	ea.Electing = e.Electing
	ea.Election = e
	ea.VMIndex = e.VMIndex
	// Build the authset
	// TODO: Check the order!

	e.LogPrintf("election", "NewElectionAdapter")
	elections.CheckAuthSetsMatch("NewElectionAdapter", e, e.State.(*state.State))

	authset := primitives.NewAuthSet()
	for _, f := range ea.Election.Federated {
		authset.AddHash(f.GetChainID(), 1)
	}

	for _, a := range buildPriorityOrder(ea.Election.Audit, dbHash, e.Minute, e.VMIndex) {
		idhash := primitives2.NewHash(a[:])
		authset.AddHash(idhash, 0)
		ea.AuditServerList = append(ea.AuditServerList, idhash)
	}

	ea.SimulatedElection = election.NewElection(primitives.Identity(e.State.GetIdentityChainID().Fixed()), *authset)
	ea.SimulatedElection.AddDisplay(nil)

	return ea
}

// Execute will:
// 	take in a message
// 	convert it to the adapted message
//	convert returned message to imsg
//	return
func (ea *ElectionAdapter) Execute(msg interfaces.IMsg) interfaces.IMsg {
	ea.Election.LogMessage("election", fmt.Sprintf("adapter_exec %d", ea.Electing), msg)

	if ea.ElectionProcessed {
		return nil // If that election is complete, just return
	}

	simmessage := ea.adaptMessage(msg)
	if simmessage == nil {
		// TODO: Handle error case
		panic("Simessage is nil")
		//return nil
	}

	//var from interfaces.IHash
	//
	//switch msg.(type) {
	//case *FedVoteVolunteerMsg:
	//	from = msg.(*FedVoteVolunteerMsg).ServerID
	//case *FedVoteProposalMsg:
	//	from = msg.(*FedVoteProposalMsg).Signer
	//case *FedVoteLevelMsg:
	//	from = msg.(*FedVoteLevelMsg).Signer
	//}

	// fmt.Printf("SimExecute In :: %s <- %s BY %x\n", ea.Election.State.GetFactomNodeName(), ea.SimulatedElection.Display.FormatMessage(simmessage), from.Fixed())

	// The second arg does not matter for our purposes
	resp, _ := ea.SimulatedElection.Execute(simmessage, 0)

	// All responses are unique and generated by us
	if resp != nil {
		// fmt.Printf("SimExecute Out :: %s -> %s BY %x\n", ea.Election.State.GetFactomNodeName(), ea.SimulatedElection.Display.FormatMessage(resp), ea.Election.State.GetIdentityChainID().Fixed())
		expandedResp := ea.expandMyMessage(resp).(interfaces.Signable)
		// Sign it!
		err := expandedResp.Sign(ea.Election.State)
		if err != nil {
			// TODO: Panic?
			panic(err)
		}
		ea.Election.LogMessage("election", fmt.Sprintf("return %d", ea.Electing), expandedResp.(interfaces.IMsg))
		return expandedResp.(interfaces.IMsg)
	}

	return nil
}

func (ea *ElectionAdapter) expandMyMessage(msg imessage.IMessage) interfaces.IMsg {
	switch msg.(type) {
	case *messages.VoteMessage:
		sim := msg.(*messages.VoteMessage)
		vol, ok := ea.Volunteers[sim.Volunteer.Signer]
		if !ok {
			panic("We should always have the volunteer message here")
		}

		p := NewFedProposalMsg(ea.Election.FedID, *vol)
		p.InitFields(ea.Election)
		p.Signer = ea.Election.State.GetIdentityChainID()

		// TODO: Set Message type
		p.TypeMsg = 0x00

		return p
	case *messages.LeaderLevelMessage:
		sim := msg.(*messages.LeaderLevelMessage)
		vol, ok := ea.Volunteers[sim.VolunteerMessage.Signer]
		if !ok {
			panic("We should always have the volunteer message here")
		}

		l := NewFedVoteLevelMessage(ea.Election.FedID, *vol)
		l.Signer = ea.Election.State.GetIdentityChainID()
		l.Level = uint32(sim.Level)
		l.Rank = uint32(sim.Rank)
		l.Committed = sim.Committed
		l.EOMFrom.SetBytes(sim.EOMFrom[:])

		for _, j := range sim.Justification {
			just := ea.expandGeneral(&j)
			if just != nil {
				// Only keep one level, so clear the justification before I include it.
				llm := just.(*FedVoteLevelMsg)
				llm.Justification = llm.Justification[:0]
				l.Justification = append(l.Justification, llm)
			}
		}

		// Just throw prev in here for now
		//prev := ea.expandGeneral(sim.PreviousVote)
		//if prev != nil {
		//	l.Justification = append(l.Justification, ea.expandGeneral(sim.PreviousVote))
		//}

		for _, j := range sim.VoteMessages {
			just := ea.expandGeneral(j)
			if just != nil {
				// TODO: Clear level to ensure just 1 level deep?
				l.Justification = append(l.Justification, just)
			}
		}

		l.InitFields(ea.Election)

		// TODO: Set Message type
		l.TypeMsg = 0x00

		ea.tagMessage(l)
		return l
	}
	// TODO: Handle error
	panic("All messages should be handled")
	//return nil
}

/***
 *
 * Expanding a message goes from simulation --> factomd
 * 	Only works for messages NOT generated by 'I'
 *
 */

func (ea *ElectionAdapter) expandMessage(msg imessage.IMessage) interfaces.IMsg {
	return ea.expandGeneral(msg)
}

func (ea *ElectionAdapter) expandGeneral(msg imessage.IMessage) interfaces.IMsg {
	tagable, ok := msg.(imessage.Taggable)
	if !ok {
		return nil
	}
	expandedGeneral, ok := ea.taggedMessages[tagable.Tag()]
	if !ok {
		return nil
	}

	return expandedGeneral
}

/***
 *
 * Adapting a message goes from factomd --> simulation
 *
 */

func (ea *ElectionAdapter) adaptMessage(msg interfaces.IMsg) imessage.IMessage {
	switch msg.(type) {
	case *FedVoteVolunteerMsg:
		return ea.adaptVolunteerMessage(msg.(*FedVoteVolunteerMsg))
	case *FedVoteProposalMsg:
		return ea.adaptVoteMessage(msg.(*FedVoteProposalMsg))
	case *FedVoteLevelMsg:
		return ea.adaptLevelMessage(msg.(*FedVoteLevelMsg), false)
	}

	return nil
}

func (ea *ElectionAdapter) adaptVolunteerMessage(msg *FedVoteVolunteerMsg) *messages.VolunteerMessage {
	ea.tagMessage(msg)

	vol := msg.ServerID.Fixed()
	volid := primitives.Identity(vol)
	volmsg := messages.NewVolunteerMessageWithoutEOM(volid)
	volmsg.TagMessage(msg.GetMsgHash().Fixed())
	return &volmsg
}

func (ea *ElectionAdapter) adaptVoteMessage(msg *FedVoteProposalMsg) *messages.VoteMessage {
	ea.tagMessage(msg)

	volmsg := ea.adaptVolunteerMessage(&msg.Volunteer)
	vote := messages.NewVoteMessage(*volmsg, primitives.Identity(msg.Signer.Fixed()))
	vote.TagMessage(msg.GetMsgHash().Fixed())
	return &vote
}

// adaptLevelMessage
// To stop possible infinite recursive behavior, only adapt the first level of justifications
func (ea *ElectionAdapter) adaptLevelMessage(msg *FedVoteLevelMsg, single bool) *messages.LeaderLevelMessage {
	ea.tagMessage(msg)

	volmsg := ea.adaptVolunteerMessage(&msg.Volunteer)
	ll := messages.NewLeaderLevelMessage(primitives.Identity(msg.Signer.Fixed()), int(msg.Rank), int(msg.Level), *volmsg)
	ll.TagMessage(msg.GetMsgHash().Fixed())
	ll.VolunteerPriority = ea.SimulatedElection.GetVolunteerPriority(volmsg.Signer)
	ll.Committed = msg.Committed
	ll.EOMFrom = msg.EOMFrom.Fixed()

	if !single {
		for _, m := range msg.Justification {
			switch m.(type) {
			case *FedVoteProposalMsg:
				p := ea.adaptVoteMessage(m.(*FedVoteProposalMsg))
				if p != nil {
					ll.VoteMessages = append(ll.VoteMessages, p)
				}
			case *FedVoteLevelMsg:
				l := ea.adaptLevelMessage(m.(*FedVoteLevelMsg), true)
				if l != nil {
					ll.Justification = append(ll.Justification, *l)
				}
			}
		}
	}

	return &ll
}

/*************/

// tagMessage is called on all adapted messages.
func (ea *ElectionAdapter) tagMessage(msg interfaces.IMsg) {
	ea.taggedMessages[msg.GetMsgHash().Fixed()] = msg
	ea.saveVolunteer(msg)
}

func (ea *ElectionAdapter) saveVolunteer(msg interfaces.IMsg) {
	switch msg.(type) {
	case *FedVoteVolunteerMsg:
		raw := msg.(*FedVoteVolunteerMsg)
		if _, ok := ea.Volunteers[raw.ServerID.Fixed()]; !ok {
			ea.Volunteers[raw.ServerID.Fixed()] = raw
		}
	case *FedVoteProposalMsg:
		raw := msg.(*FedVoteProposalMsg)
		if _, ok := ea.Volunteers[raw.Volunteer.ServerID.Fixed()]; !ok {
			ea.Volunteers[raw.Volunteer.ServerID.Fixed()] = &raw.Volunteer
		}
	case *FedVoteLevelMsg:
		raw := msg.(*FedVoteLevelMsg)
		if _, ok := ea.Volunteers[raw.Volunteer.ServerID.Fixed()]; !ok {
			ea.Volunteers[raw.Volunteer.ServerID.Fixed()] = &raw.Volunteer
		}
	}
}

func (ea *ElectionAdapter) GetDBHeight() int {
	return ea.DBHeight
}

func (ea *ElectionAdapter) GetVMIndex() int {
	return ea.VMIndex
}

func (ea *ElectionAdapter) GetMinute() int {
	return ea.Minute
}

func (ea *ElectionAdapter) GetElecting() int {
	return ea.Electing
}

func (ea *ElectionAdapter) IsObserver() bool {
	return ea.SimulatedElection.Observer
}

func (ea *ElectionAdapter) SetObserver(o bool) {
	ea.SimulatedElection.SetObserver(o)
}

func (ea *ElectionAdapter) IsElectionProcessed() bool {
	return ea.ElectionProcessed
}

func (ea *ElectionAdapter) SetElectionProcessed(swapped bool) {
	ea.ElectionProcessed = swapped
}

func (ea *ElectionAdapter) IsStateProcessed() bool {
	return ea.StateProcessed
}

func (ea *ElectionAdapter) SetStateProcessed(swapped bool) {
	ea.StateProcessed = swapped
}
