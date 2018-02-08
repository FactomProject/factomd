package messages

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
		var vote VoteMessage
		for _, v := range md.MajorityVotes {
			vote = v
			break
		}
		return &vote.Volunteer
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
