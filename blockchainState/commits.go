// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (bs *BlockchainState) HasFreeCommit(h interfaces.IHash) bool {
	pc, ok := bs.PendingCommits[h.String()]
	if ok == false {
		return false
	}
	switch h.String() {
	case "":
		fmt.Printf("Missing commit - %v\n", pc.String())
		break
	}

	return pc.HasFreeCommit()
}

func (bs *BlockchainState) PopCommit(h interfaces.IHash) error {
	pc, ok := bs.PendingCommits[h.String()]
	if ok == false {
		return fmt.Errorf("No commits found")
	}
	return pc.PopCommit(bs.DBlockHeight)
}

func (bs *BlockchainState) PushCommit(entryHash interfaces.IHash, commitTxID interfaces.IHash) {
	if bs.PendingCommits[entryHash.String()] == nil {
		bs.PendingCommits[entryHash.String()] = new(PendingCommit)
	}
	if MES.IsEntryMissing(entryHash.String()) {
		MES.FoundMissing(entryHash.String(), commitTxID.String(), bs.DBlockHead.KeyMR.String(), bs.DBlockHeight)
		return
	}
	bs.PendingCommits[entryHash.String()].PushCommit(commitTxID, bs.DBlockHeight)
}

func (bs *BlockchainState) ClearExpiredCommits() error {
	for k, v := range bs.PendingCommits {
		v.ClearExpiredCommits(bs.DBlockHeight, bs.IsMainNet())
		if v.HasFreeCommit() == false {
			delete(bs.PendingCommits, k)
		}
	}
	return nil
}

type PendingCommit struct {
	Commits []SingleCommit
}

func (pc *PendingCommit) String() string {
	str, _ := primitives.EncodeJSONString(pc)
	return str
}

func (pc *PendingCommit) HasFreeCommit() bool {
	if len(pc.Commits) > 0 {
		return true
	}
	return false
}

func (pc *PendingCommit) PopCommit(dblockHeight uint32) error {
	if len(pc.Commits) == 0 {
		return fmt.Errorf("No commits found")
	}
	if int(dblockHeight-pc.Commits[0].DBlockHeight) > LatestReveal {
		LatestReveal = int(dblockHeight - pc.Commits[0].DBlockHeight)
	}
	pc.Commits = pc.Commits[1:]
	return nil
}

func (pc *PendingCommit) PushCommit(commitTxID interfaces.IHash, dblockHeight uint32) {
	pc.Commits = append(pc.Commits, SingleCommit{DBlockHeight: dblockHeight, CommitTxID: commitTxID.String()})
}

func (pc *PendingCommit) ClearExpiredCommits(dblockHeight uint32, mainNet bool) {
	for {
		if len(pc.Commits) == 0 {
			return
		}
		if mainNet {
			if dblockHeight < M2SWITCHHEIGHT {
				if pc.Commits[0].DBlockHeight+COMMITEXPIRATIONM1 < dblockHeight {
					pc.Commits = pc.Commits[1:]
					Expired++
				} else {
					return
				}
			} else {
				if pc.Commits[0].DBlockHeight+COMMITEXPIRATIONM2 < dblockHeight {
					pc.Commits = pc.Commits[1:]
					Expired++
				} else {
					return
				}
			}
		} else {
			//Non-MainNet
			if pc.Commits[0].DBlockHeight+COMMITEXPIRATIONM2 < dblockHeight {
				pc.Commits = pc.Commits[1:]
				Expired++
			} else {
				return
			}
		}
	}
}

type SingleCommit struct {
	DBlockHeight uint32
	CommitTxID   string
}
