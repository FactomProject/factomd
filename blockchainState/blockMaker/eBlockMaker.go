// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"sort"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (bm *BlockMaker) BuildEBlocks() ([]interfaces.IEntryBlock, error) {
	sortedEntries := map[string][]interfaces.IEntry{}
	for _, v := range bm.ProcessedEBEntries {
		sortedEntries[v.GetChainID().String()] = append(sortedEntries[v.GetChainID().String()], v)
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
		eb.GetHeader().SetChainID(entries[0].GetChainID())
		head := bm.BState.GetEBlockHead(entries[0].GetChainID().String())
		eb.GetHeader().SetPrevKeyMR(head.KeyMR)
		eb.GetHeader().SetPrevFullHash(head.Hash)
		//...
		eBlocks = append(eBlocks, eb)
	}
	return eBlocks, nil
}

func (bm *BlockMaker) ProcessEBEntry(e interfaces.IEntry) error {
	return nil
}
