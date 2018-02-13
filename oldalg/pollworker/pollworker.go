package pollworker

import (
	"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/util"
)

// PollWorker controls only 1 minute, but all VMs in that minute. His job
// is to get EOMs, or publishes if an EOM is not found
type PollWorker struct {
	// Each index is a VM
	Elections []*election.Election
	// Come from elections
	ElectionPublishes []*messages.PublishMessage
	// Come from the original leader
	TrumpEOMs []*messages.EomMessage

	// The authset is fixed, however we can have an audit that signs in place of
	// a fed who is faulted
	primitives.AuthSet
	// We need our minute and our height
	primitives.MinuteLocation

	Self primitives.Identity
}

func NewPollWorker(self primitives.Identity, a primitives.AuthSet, location primitives.MinuteLocation) *PollWorker {
	p := new(PollWorker)
	p.Self = self
	p.AuthSet = a
	p.MinuteLocation = location

	// A VM per leader
	numFeds := len(p.GetFeds())
	p.Elections = make([]*election.Election, numFeds)
	p.TrumpEOMs = make([]*messages.EomMessage, numFeds)
	p.ElectionPublishes = make([]*messages.PublishMessage, numFeds)

	return p
}

func (p *PollWorker) Execute(msg imessage.IMessage) []imessage.IMessage {
	response := []imessage.IMessage{}

	// If we have a trump EOM or an EOM, we must always return that
	vm := p.GetVMForMsg(msg)
	if vm == -1 {
		return response
	}

	// Trump always wins
	if p.TrumpEOMs[vm] != nil {
		return imessage.MakeMessageArrayFromArray(response, *p.TrumpEOMs[vm])
	}

	// Election trump does not always win
	if p.ElectionPublishes[vm] != nil {
		response = imessage.MakeMessageArrayFromArray(response, *p.ElectionPublishes[vm])
	}

	switch msg.(type) {
	case messages.EomMessage:
		// This has come from the process list, as it came from above not below.
		eom := msg.(messages.EomMessage)
		if eom.MinuteLocation != p.MinuteLocation {
			// Not the minute we care about
			return response
		}

		vm := p.VMForIdentity(eom.Signer, p.MinuteLocation)
		if vm != eom.Vm {
			// This EOM must be the leader, however for some reason it's not matching. Toss it
			return response
		}

		p.TrumpEOMs[vm] = &eom
		return imessage.MakeMessageArray(eom)
	case messages.FaultMsg:
		fault := msg.(messages.FaultMsg)
		vm := p.VMForIdentity(fault.Replacing, p.MinuteLocation)
		response = append(response, p.executeInElection(vm).ExecuteMsg(msg)...)
		return response
	default:
		vm := p.GetVMForMsg(msg)
		response = append(response, p.executeInElection(vm).ExecuteMsg(msg)...)
		return response
	}
	return response
}

// acceptMsg takes messages from the elections and checks them before sending out
// If they have a resolution, we need to se out trumpEOM if it has not been set by
// a round 0.
func (p *PollWorker) acceptMsg(msgs []imessage.IMessage) []imessage.IMessage {
	for _, msg := range msgs {
		switch msg.(type) {
		case messages.PublishMessage:
			pub := msg.(messages.PublishMessage)
			vol := messages.GetVolunteerMsg(pub)
			p.ElectionPublishes[p.VMForIdentity(vol.Replacing, p.MinuteLocation)] = &pub
		}
	}

	return msgs
}

func (p *PollWorker) GetVMForMsg(msg imessage.IMessage) int {
	return util.GetVMForMsg(msg, p.AuthSet, p.MinuteLocation)
}

// executeInElection guarantees the election exists
func (p *PollWorker) executeInElection(vm int) *election.Election {
	e := p.Elections[vm]
	if e == nil {
		e = election.NewElection(p.Self, p.AuthSet, primitives.ProcessListLocation{vm, p.MinuteLocation})
		p.Elections[vm] = e
	}

	return e
}
