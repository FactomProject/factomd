package util

import (
	"fmt"
	"runtime"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
)

// a simple file/line trace function, with optional comment(s)
func Trace(params ...string) {
	log.Printf("##")

	if 0 < len(params) {
		for i := range params {
			log.Printf(" %s", params[i])
		}
		log.Printf(" #### ")
	} else {
		log.Printf(" ")
	}

	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])

	tutc := time.Now().UTC()
	timestamp := tutc.Format("2006-01-02.15:04:05")

	log.Printf("TRACE: %s line %d %s file: %s\n", timestamp, line, f.Name(), file)
}

// Calculate the entry credits needed for the entry
func EntryCost(b []byte) (uint8, error) {
	// caulculaate the length exluding the header size 35 for Milestone 1
	l := len(b) - 35

	if l > 10240 {
		return 10, fmt.Errorf("Entry cannot be larger than 10KB")
	}

	// n is the capacity of the entry payment in KB
	r := l % 1024
	n := uint8(l / 1024)

	if r > 0 {
		n += 1
	}

	if n < 1 {
		n = 1
	}

	return n, nil
}

func IsInPendingEntryList(list []interfaces.IPendingEntry, entry interfaces.IPendingEntry) bool {
	if len(list) == 0 {
		return false
	}
	for k, ent := range list {

		if entry.ChainID != nil {

			if entry.EntryHash != nil {

				if entry.EntryHash.IsSameAs(ent.EntryHash) {
					if list[k].ChainID == nil {
						// this is the only time we have these two data items at the same time.  if you already have a chain commit, you don't know the chain on it
						// update the chain IDless entry with a chain ID instead of adding another that knows chainid and entryhash
						list[k].ChainID = entry.ChainID
					}
					if entry.ChainID.IsSameAs(ent.ChainID) {
						return true
					}
				}
			}
		}
	}
	return false
}
