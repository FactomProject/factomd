package controller

import (
	"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/messages"
)

// Router helps keep route patterns for returning messages
type Router struct {
	Elections  []*election.Election
	Volunteers []*messages.VolunteerMessage
}
