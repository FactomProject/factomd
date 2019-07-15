package servertype

import (
	"github.com/FactomProject/factomd/state"
)

type ServerType string

const (
	Follower        ServerType = "follower"
	AuditServer     ServerType = "audit server"
	FederatedServer ServerType = "federated server"
)

func GetServerType(list *state.ProcessList, state *state.State) ServerType {
	if state.Leader {
		return FederatedServer
	}

	foundAudit, _ := list.GetAuditServerIndexHash(state.GetIdentityChainID())
	if foundAudit {
		return AuditServer
	}

	return Follower
}
