// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"sort"
	"strings"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (bm *BlockMaker) BuildEBlocks() ([]interfaces.IEntryBlock, error) {
	sortedEntries := map[string][]*EBlockEntry{}
	sort.Sort(EBlockEntryByMinute(bm.ProcessedEBEntries))
	for _, v := range bm.ProcessedEBEntries {
		sortedEntries[v.Entry.GetChainID().String()] = append(sortedEntries[v.Entry.GetChainID().String()], v)
	}
	keys := []string{}
	for k := range sortedEntries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	eBlocks := []interfaces.IEntryBlock{}
	for _, k := range keys {
		entries := sortedEntries[k]
		if len(entries) == 0 {
			continue
		}
		eb := entryBlock.NewEBlock()
		eb.GetHeader().SetChainID(entries[0].Entry.GetChainID())
		head := bm.BState.GetEBlockHead(entries[0].Entry.GetChainID().String())
		eb.GetHeader().SetPrevKeyMR(head.KeyMR)
		eb.GetHeader().SetPrevFullHash(head.Hash)
		eb.GetHeader().SetDBHeight(bm.BState.DBlockHeight + 1)

		minute := entries[0].Minute
		for _, v := range entries {
			if v.Minute != minute {
				eb.AddEndOfMinuteMarker(uint8(minute + 1))
				minute = v.Minute
			}
			eb.AddEBEntry(v.Entry)
		}
		eb.AddEndOfMinuteMarker(uint8(minute + 1))
		eBlocks = append(eBlocks, eb)
	}
	return eBlocks, nil
}

func (bm *BlockMaker) ProcessEBEntry(e interfaces.IEntry) error {
	//TODO: handle minute markers
	//TODO: do
	ebe := new(EBlockEntry)
	ebe.Entry = e
	ebe.Minute = bm.CurrentMinute
	bm.ProcessedEBEntries = append(bm.ProcessedEBEntries, ebe)
	return nil
}

type EBlockEntry struct {
	Entry  interfaces.IEntry
	Minute int
}

type EBlockEntryByMinute []*EBlockEntry

func (f EBlockEntryByMinute) Len() int {
	return len(f)
}
func (f EBlockEntryByMinute) Less(i, j int) bool {
	if f[i].Minute < f[j].Minute {
		return true
	}
	if f[i].Minute > f[j].Minute {
		return false
	}
	return strings.Compare(f[i].Entry.GetHash().String(), f[j].Entry.GetHash().String()) < 0
}
func (f EBlockEntryByMinute) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
