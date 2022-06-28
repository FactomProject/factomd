package code

import (
	"time"

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

			// When syncing with the network, the entry isn't necessarily ready
			// from the database immediately.  So go into a loop where if
			// the entry isn't ready (which should be there), we pause and try
			// again
			var entry interfaces.IEBEntry
			for {
				entry, err = DB.FetchEntry(e)
				if err != nil {
					panic("Missing Entry")
				}
				if entry != nil {
					break
				}
				time.Sleep(10*time.Millisecond) // Don't go crazy CPU wise if an Entry isn't ready
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
