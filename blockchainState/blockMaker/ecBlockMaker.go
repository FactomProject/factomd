// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (bm *BlockMaker) BuildECBlock() (interfaces.IEntryCreditBlock, error) {
	sort.Sort(ECBlockEntryByMinute(bm.ProcessedECBEntries))

	ecBlock := entryCreditBlock.NewECBlock()
	ecBlock.GetHeader().SetPrevFullHash(bm.BState.ECBlockHead.Hash)
	ecBlock.GetHeader().SetPrevHeaderHash(bm.BState.ECBlockHead.KeyMR)
	ecBlock.GetHeader().SetDBHeight(bm.BState.ECBlockHead.Height + 1)

	minute := 0
	for _, v := range bm.ProcessedECBEntries {
		for ; minute < v.Minute; minute++ {
			e := entryCreditBlock.NewMinuteNumber(uint8(minute + 1))
			ecBlock.GetBody().AddEntry(e)
		}
		ecBlock.GetBody().AddEntry(v.Entry)
	}
	for ; minute < 9; minute++ {
		e := entryCreditBlock.NewMinuteNumber(uint8(minute + 1))
		ecBlock.GetBody().AddEntry(e)
	}

	err := ecBlock.BuildHeader()
	if err != nil {
		return nil, err
	}

	return ecBlock, nil
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
