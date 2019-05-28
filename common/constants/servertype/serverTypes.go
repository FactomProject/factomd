package servertype

import (
	"github.com/FactomProject/factomd/state"
)

type ServerType string

const (
	Follower    ServerType = "follower"
	AuditServer ServerType = "audit server"
	Leader      ServerType = "leader"
)

func GetServerType(list *state.ProcessList, state *state.State) ServerType {
	if state.Leader {
		return Leader
	}

	foundAudit, _ := list.GetAuditServerIndexHash(state.GetIdentityChainID())
	if foundAudit {
		return AuditServer
	}

	return Follower
}
