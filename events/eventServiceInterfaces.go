package events

import (
	"github.com/PaulSnow/factom2d/common/constants/runstate"
	"github.com/PaulSnow/factom2d/common/interfaces"
)

type StateEventServices interface {
	GetRunState() runstate.RunState
	GetIdentityChainID() interfaces.IHash
	IsRunLeader() bool
	GetEventService() EventService
}
