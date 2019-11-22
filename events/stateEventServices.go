package events

import "github.com/FactomProject/factomd/events/eventservices"

type IStateEventServices interface {
	GetEventsService() eventservices.EventService

	IsRunLeader() bool
}
