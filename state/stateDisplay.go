package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/modules/event"
	"strings"
)

func (s *State) stateUpdate() *events.StateUpdate {
	fnodes := []*State{s}
	nodesSummary := nodesSummary(fnodes)
	summary := fmt.Sprintf("===SummaryStart===%s \n%s===SummaryEnd===\n", s.ShortString(), nodesSummary)

	identitiesDetails := identitiesDetails(s.IdentityControl.GetSortedIdentities())
	authoritiesDetails := authoritiesDetails(s.IdentityControl.GetSortedAuthorities())

	return &events.StateUpdate{
		NodeTime:           s.ProcessTime,
		LeaderHeight:       s.LLeaderHeight,
		Summary:            summary,
		IdentitiesDetails:  identitiesDetails,
		AuthoritiesDetails: authoritiesDetails,
	}
}

// Data Dump String Creation
func nodesSummary(fnodes []*State) string {
	var nodesSummary strings.Builder
	var nodes, review, holding, acks, msgQueue, prioritizedMsgQueue, inMsgQueue, apiQueue, ackQueue, timerMsgQueue, networkOutMsgQueue, networkInvalidMsgQueue string
	for i, f := range fnodes {
		nodes = fmt.Sprintf("%s %3d", nodes, i)
		review = fmt.Sprintf("%s %3d", review, len(f.XReview))
		holding = fmt.Sprintf(" %3d", len(f.Holding))
		acks = fmt.Sprintf(" %3d", len(f.Acks))
		msgQueue = fmt.Sprintf(" %3d", len(f.MsgQueue()))
		prioritizedMsgQueue = fmt.Sprintf(" %3d", len(f.PrioritizedMsgQueue()))
		inMsgQueue = fmt.Sprintf(" %3d", f.InMsgQueue().Length())
		apiQueue = fmt.Sprintf(" %3d", f.APIQueue().Length())
		ackQueue = fmt.Sprintf(" %3d", len(f.AckQueue()))
		timerMsgQueue = fmt.Sprintf(" %3d", len(f.TimerMsgQueue()))
		networkOutMsgQueue = fmt.Sprintf(" %3d", f.NetworkOutMsgQueue().Length())
		networkInvalidMsgQueue = fmt.Sprintf(" %3d", len(f.NetworkInvalidMsgQueue()))

	}

	format := "%22s%s\n"
	_, _ = fmt.Fprintf(&nodesSummary, format, "", nodes)
	_, _ = fmt.Fprintf(&nodesSummary, format, "Review", review)
	_, _ = fmt.Fprintf(&nodesSummary, format, "Holding", holding)
	_, _ = fmt.Fprintf(&nodesSummary, format, "Acks", acks)
	_, _ = fmt.Fprintf(&nodesSummary, format, "MsgQueue", msgQueue)
	_, _ = fmt.Fprintf(&nodesSummary, format, "PrioritizedMsgQueue", prioritizedMsgQueue)
	_, _ = fmt.Fprintf(&nodesSummary, format, "InMsgQueue", inMsgQueue)
	_, _ = fmt.Fprintf(&nodesSummary, format, "APIQueue", apiQueue)
	_, _ = fmt.Fprintf(&nodesSummary, format, "AckQueue", ackQueue)
	_, _ = fmt.Fprintf(&nodesSummary, format, "TimerMsgQueue", timerMsgQueue)
	_, _ = fmt.Fprintf(&nodesSummary, format, "NetworkOutMsgQueue", networkOutMsgQueue)
	_, _ = fmt.Fprintf(&nodesSummary, format, "NetworkInvalidMsgQueue", networkInvalidMsgQueue)

	return nodesSummary.String()
}

func identitiesDetails(identities []*identity.Identity) string {
	var details strings.Builder
	_, _ = fmt.Fprintf(&details, "=== Identity List ===   Total: %d Displaying: All\n", len(identities))
	for num, i := range identities {
		_, _ = fmt.Fprintf(&details, "------------------------------------%d---------------------------------------\n", num)
		_, _ = fmt.Fprintf(&details, "Server Status: %s\n", constants.IdentityStatusString(i.Status))
		_, _ = fmt.Fprintf(&details, "Synced Status: ID[%t] MG[%t]\n", i.IdentityChainSync.Synced(), i.ManagementChainSync.Synced())
		_, _ = fmt.Fprintf(&details, "Identity Chain: %s (C:%d R:%d)\n", i.IdentityChainID.String(), i.IdentityCreated, i.IdentityRegistered)
		_, _ = fmt.Fprintf(&details, "Management Chain: %s (C:%d R:%d)\n", i.ManagementChainID.String(), i.ManagementCreated, i.ManagementRegistered)
		_, _ = fmt.Fprintf(&details, "Matryoshka Hash: %s\n", i.MatryoshkaHash)
		_, _ = fmt.Fprintf(&details, "Key 1: %s\n", i.Keys[0])
		_, _ = fmt.Fprintf(&details, "Key 2: %s\n", i.Keys[1])
		_, _ = fmt.Fprintf(&details, "Key 3: %s\n", i.Keys[2])
		_, _ = fmt.Fprintf(&details, "Key 4: %s\n", i.Keys[3])
		_, _ = fmt.Fprintf(&details, "Signing Key: %s\n", i.SigningKey)
		_, _ = fmt.Fprintf(&details, "Coinbase Address: %s\n", i.GetCoinbaseHumanReadable())
		_, _ = fmt.Fprintf(&details, "Efficiency: %s&#37;\n", primitives.EfficiencyToString(i.Efficiency))

		for _, a := range i.AnchorKeys {
			_, _ = fmt.Fprintf(&details, "Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey)
		}
		_, _ = fmt.Fprintf(&details, "ID Eblock Syncing: Current: %d  Target: %d\n", i.IdentityChainSync.Current.DBHeight, i.IdentityChainSync.Target.DBHeight)
		_, _ = fmt.Fprintf(&details, "MG Eblock Syncing: Current: %d  Target: %d\n", i.ManagementChainSync.Current.DBHeight, i.ManagementChainSync.Target.DBHeight)
	}

	return details.String()
}

func authoritiesDetails(authorities []interfaces.IAuthority) string {
	var details strings.Builder
	_, _ = fmt.Fprintf(&details, "=== Authority List ===   Total: %d Displaying: All\n", len(authorities))
	for num, i := range authorities {
		if authority, ok := i.(*identity.Authority); ok {
			_, _ = fmt.Fprintf(&details, "------------------------------------%d---------------------------------------\n", num)
			_, _ = fmt.Fprintf(&details, "Server Status: %s\n", constants.IdentityStatusString(authority.Status))
			_, _ = fmt.Fprintf(&details, "Identity Chain: %s\n", authority.AuthorityChainID)
			_, _ = fmt.Fprintf(&details, "Management Chain: %s\n", authority.ManagementChainID)
			_, _ = fmt.Fprintf(&details, "Matryoshka Hash: %s\n", authority.MatryoshkaHash)
			_, _ = fmt.Fprintf(&details, "Signing Key: %s\n", authority.SigningKey.String())
			_, _ = fmt.Fprintf(&details, "Coinbase Address: %s\n", authority.GetCoinbaseHumanReadable())
			_, _ = fmt.Fprintf(&details, "Efficiency: %d\n", authority.Efficiency)

			for _, a := range authority.AnchorKeys {
				_, _ = fmt.Fprintf(&details, "Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey)
			}
		}
	}
	return details.String()
}
