package identity

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/identityEntries"
)

// CoinbaseCancelManager handles keeping track of coinbase cancel signals
// in identity chains. Specifically keeps track of the tallies, then can determine
// if an admin block entry should be created.
type CoinbaseCancelManager struct {
	// Heights of various descriptor cancel proposals. This list will be
	// sorted, and is used for garbage collecting proposals that have failed
	ProposalsList []uint32

	// Proposals is all cancel proposals for a given descriptor height
	//		[descriptorheight][cancel index]Cancel identity entry
	Proposals map[uint32]map[uint32][]identityEntries.NewCoinbaseCancelStruct

	// Boolean indicator if it's been recorded to the admin block. We do not do this more than once
	AdminBlockRecord map[uint32]map[uint32]bool
}

func NewCoinbaseCancelManager() *CoinbaseCancelManager {
	c := new(CoinbaseCancelManager)
	c.Proposals = make(map[uint32]map[uint32][]identityEntries.NewCoinbaseCancelStruct)
	c.AdminBlockRecord = make(map[uint32]map[uint32]bool)

	return c
}

// GC is garbage collecting old proposals
//		dbheight is the current height.
func (c *CoinbaseCancelManager) GC(dbheight uint32) {
	// These are sorting in incrementing order.
	for _, h := range c.ProposalsList {
		// If you height a height that is greater than the height, break.
		// These are still valid proposals
		if h+constants.COINBASE_DECLARATION > dbheight {
			break
		}
		delete(c.Proposals, h)
		delete(c.AdminBlockRecord, h)
	}
}

// AddCancel will add a proposal to the list. It assumes the height check has already been done
func (cm *CoinbaseCancelManager) AddCancel(cc identityEntries.NewCoinbaseCancelStruct) {
	list, ok := cm.Proposals[cc.CoinbaseDescriptorHeight]
	if !ok {
		// A new height is added, we also need to insert it into our proposalsList
		cm.AddNewProposalHeight(cc.CoinbaseDescriptorHeight)
		cm.Proposals[cc.CoinbaseDescriptorHeight] = make(map[uint32][]identityEntries.NewCoinbaseCancelStruct, 0)
		cm.AdminBlockRecord[cc.CoinbaseDescriptorHeight] = make(map[uint32]bool, 0)
		list = cm.Proposals[cc.CoinbaseDescriptorHeight]
	}

	index, ok := list[cc.CoinbaseDescriptorIndex]
	if !ok {
		cm.Proposals[cc.CoinbaseDescriptorHeight][cc.CoinbaseDescriptorIndex] = []identityEntries.NewCoinbaseCancelStruct{}
		index = cm.Proposals[cc.CoinbaseDescriptorHeight][cc.CoinbaseDescriptorIndex]
	}

	cm.Proposals[cc.CoinbaseDescriptorHeight][cc.CoinbaseDescriptorIndex] = append(index, cc)
}

// CanceledOutputs will return the indices of all indices to be canceled for a given descriptor height
func (cm *CoinbaseCancelManager) CanceledOutputs(descriptorHeight uint32) []int {
	return []int{}
}

// MarkAdminBlockRecorded will mark a given index for a descriptor already canceled. This is to prevent
// a given index from being recorded multiple times
func (cm *CoinbaseCancelManager) MarkAdminBlockRecorded(descriptorHeight uint32, index uint32) {
	if _, ok := cm.AdminBlockRecord[descriptorHeight]; !ok {
		cm.AddNewProposalHeight(descriptorHeight)
		cm.AdminBlockRecord[descriptorHeight] = make(map[uint32]bool, 0)
		cm.Proposals[descriptorHeight] = make(map[uint32][]identityEntries.NewCoinbaseCancelStruct, 0)
	}

	cm.AdminBlockRecord[descriptorHeight][index] = true
}

// IsAdminBlockRecorded returns boolean if marked already recorded. Garbage collected heights
// will return false, so the caller will have to check the dbheight is valid.
func (cm *CoinbaseCancelManager) IsAdminBlockRecorded(descriptorHeight uint32, index uint32) bool {
	if list, ok := cm.AdminBlockRecord[descriptorHeight]; ok {
		if value, ok := list[index]; ok {
			return value
		}
	}
	return false
}

// AddNewProposalHeight does a insert into the sorted list. It does not use binary search as this list
// should be relatively small, and infrequently used. Only used when canceling a coinbase in the future.
func (cm *CoinbaseCancelManager) AddNewProposalHeight(descriptorHeight uint32) {
	for i := len(cm.ProposalsList) - 1; i >= 0; i-- {
		if descriptorHeight > cm.ProposalsList[i] {
			// Insert into list
			cm.ProposalsList = append(cm.ProposalsList, 0)
			copy(cm.ProposalsList[i+1:], cm.ProposalsList[i:])
			cm.ProposalsList[i] = descriptorHeight
			return
		}
	}
	cm.ProposalsList = append([]uint32{descriptorHeight}, cm.ProposalsList...)
}
