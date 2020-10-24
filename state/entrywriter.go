package state

import (
	"github.com/FactomProject/factomd/database/databaseOverlay"
)

// WriteEntries consumed the WriteEntry channel and saves entries into the database.
// Assumes entries are valid.
// Panics on database errors
func (s *State) WriteEntries() {
	for entry := range s.WriteEntry {
		if entry == nil {
			continue
		}

		if has, err := s.DB.DoesKeyExist(databaseOverlay.ENTRY, entry.GetHash().Bytes()); err != nil {
			panic(err)
		} else if !has {
			if err = s.DB.InsertEntry(entry); err != nil {
				panic(err)
			}
		}
	}
}
