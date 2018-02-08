package util

import (
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
)

func GetVMForMsg(msg imessage.IMessage, authset primitives.AuthSet, loc primitives.MinuteLocation) int {
	// If there is no volunteer msg it is not a
	vol := messages.GetVolunteerMsg(msg)
	if vol == nil {
		return -1
	}

	return authset.VMForIdentity(vol.FaultMsg.Replacing, loc)
}
