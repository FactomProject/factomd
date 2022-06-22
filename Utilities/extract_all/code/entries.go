package code

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

func ProcessEntries(DBlock interfaces.IDirectoryBlock) {

	for _, dbe := range DBlock.GetDBEntries()[3:] {
		eb, err := DB.FetchEBlock(dbe.GetKeyMR())
		if err != nil {
			panic("Bad Entry block")
		}
		data, err := eb.MarshalBinary()
		if err != nil {
			panic("Bad Entry block")
		}
		header := Header{Tag: TagEBlock, Size: uint64(len(data))}
		OutputFile.Write(header.MarshalBinary())
		OutputFile.Write(data)

		for _, e := range eb.GetEntryHashes() {
			if e.IsMinuteMarker() {
				continue
			}
			entry, err := DB.FetchEntry(e)
			if err != nil {
				panic("Missing Entry")
			}
			entryData, err := entry.MarshalBinary()
			if err != nil {
				panic("Bad Entry")
			}
			h := Header{Tag: TagEntry, Size: uint64(len(entryData))}
			OutputFile.Write(h.MarshalBinary())
			OutputFile.Write(entryData)
			EntryCnt++
		}
	}
}
