package servertype

import (
	"github.com/FactomProject/factomd/state"
)

// ServerType is a string which describes which type of server it is
type ServerType string

const (
	Follower        ServerType = "follower"         // All servers that are not audit or federated, are followers (may or may not have an identity)
	AuditServer     ServerType = "audit server"     // Audit serveres have identities
	FederatedServer ServerType = "federated server" // Federated servers have identities
)

// GetServerType returns the server's type
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
