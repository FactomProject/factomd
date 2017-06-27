// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"strings"

	"github.com/FactomProject/factomd/common/interfaces"
)

func (bm *BlockMaker) BuildECBlock() (interfaces.IEntryCreditBlock, error) {
	return nil, nil
}

func (bm *BlockMaker) ProcessECEntry(e interfaces.IECBlockEntry) error {
	ebe := new(ECBlockEntry)
	ebe.Entry = e
	ebe.Minute = bm.CurrentMinute
	if bm.BState.ProcessECEntry(e) != nil {
		bm.PendingECBEntries = append(bm.PendingECBEntries, ebe)
	} else {
		bm.ProcessedECBEntries = append(bm.ProcessedECBEntries, ebe)
	}
	return nil
}

type ECBlockEntry struct {
	Entry  interfaces.IECBlockEntry
	Minute int
}

type ECBlockEntryByMinute []*ECBlockEntry

func (f ECBlockEntryByMinute) Len() int {
	return len(f)
}
func (f ECBlockEntryByMinute) Less(i, j int) bool {
	if f[i].Minute < f[j].Minute {
		return true
	}
	if f[i].Minute > f[j].Minute {
		return false
	}
	return strings.Compare(f[i].Entry.GetHash().String(), f[j].Entry.GetHash().String()) < 0
}
func (f ECBlockEntryByMinute) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
