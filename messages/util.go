package messages

import "github.com/FactomProject/electiontesting/primitives"

func GetSigner(msg interface{}) primitives.Identity {
	switch msg.(type) {
	case VolunteerMessage:
		v := msg.(*VolunteerMessage)
		return v.Signer
	case VoteMessage:
		v := msg.(*VoteMessage)
		return v.Signer
	case LeaderLevelMessage:
		v := msg.(*LeaderLevelMessage)
		return v.Signer
	}
	return primitives.NewIdentityFromInt(-1)
}

// GetVolunteerMsg gets the volunteer message from an election message, used to determine round
func GetVolunteerMsg(msg interface{}) *VolunteerMessage {
	switch msg.(type) {
	case VolunteerMessage:
		v := msg.(VolunteerMessage)
		return &v
	case VoteMessage:
		v := msg.(VoteMessage)
		return &v.Volunteer
	case MajorityDecisionMessage:
		md := msg.(MajorityDecisionMessage)
		// ??? what's this
		/*
			var vote SignedMessage
			for _, v := range md.MajorityVotes {
				vote = v
				break
			}
		*/
		return &md.Volunteer
	case InsistMessage:
		insist := msg.(InsistMessage)
		var md MajorityDecisionMessage

		for _, m := range insist.MajorityMajorityDecisions {
			md = m
			break
		}
		return GetVolunteerMsg(md)
	case IAckMessage:
		iack := msg.(IAckMessage)
		return GetVolunteerMsg(iack.Insist)
	case PublishMessage:
		p := msg.(PublishMessage)
		return GetVolunteerMsg(p.Insist)
	}
	return nil
}
