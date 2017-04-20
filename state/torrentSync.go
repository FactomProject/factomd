package state

import (
	"fmt"
	"log"
	"time"
)

// StartTorrentSyncing is an endless loop that uses torrents to sync missing blocks
// It will grab any block higher than the highest dblock saved in the database up
// to the highest known block.
func (s *State) StartTorrentSyncing() error {
	if !s.UsingTorrent() {
		return fmt.Errorf("State is not using torrents, yet torrent sync was called")
	}

	// Upload we have done up to
	var done uint32 = 0
	for {
		// Leaders do not need to sync torrents, they need to upload
		if s.IsLeader() {
			// If we have not uploaded a height we have completed, increment done and upload
			if done < s.EntryDBHeightComplete {
				for done < s.EntryDBHeightComplete {
					done++
					s.UploadDBState(done)
				}
			} else {
				// If we did not just launch, and we are synced, and uploaded --> Long sleep
				if s.EntryDBHeightComplete > 0 && s.GetHighestKnownBlock() == s.EntryDBHeightComplete {
					time.Sleep(30 * time.Second)
				}
				// Short sleep otherwise, still loading some from disk
				time.Sleep(5 * time.Second)
			}
			continue
		}

		// We can adjust the sleep at the end depending on what we do in the loop
		// this pass
		rightDuration := time.Duration(time.Second * 1)

		// What is the database at
		dblock, err := s.DB.FetchDBlockHead()
		if err != nil || dblock == nil {
			if err != nil {
				log.Printf("[TorrentSync] Error while retrieving dblock head, %s", err.Error())
			}
			time.Sleep(5 * time.Second) // To prevent error spam
			continue
		}

		// Range of heights to request
		lower := dblock.GetDatabaseHeight()
		upper := s.GetHighestKnownBlock()

		if upper-(BATCH_SIZE*2) < lower {
			lower = s.EntryDBHeightComplete
		}

		// If the network is at block 0, we aren't on the network
		if upper == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		// Synced up, sleep for awhile
		if lower == upper || upper-BATCH_SIZE < lower {
			time.Sleep(20 * time.Second)
			continue
		}

		// Prometheus
		stateTorrentSyncingLower.Set(float64(lower))
		stateTorrentSyncingUpper.Set(float64(upper))

		// How many blocks ahead of the current we should request from the plugin
		allowed := 5000

		// What is the end height we request
		max := lower + uint32(allowed)
		if upper < max {
			rightDuration = time.Duration(5 * time.Second)
			max = upper
		}

		var u uint32 = 0
		// The torrent plugin handles dealing with lots of heights. It has it's own queueing system, so
		// we can spam and repeat heights
	RequestLoop:
		for u = lower; u < max; u++ {
			if (upper - BATCH_SIZE) < u {
				break RequestLoop // This means we hit the highest torrent height
			}
			// Plugin handles repeat requests
			err := s.DBStateManager.RetrieveDBStateByHeight(u)
			if err != nil {
				if s.DBStateManager.Alive() == nil { // Some errors come from a plugin crash (like when you ctrl+c)
					log.Printf("[TorrentSync] Error while retrieving height %d by torrent, %s", u, err.Error())
				} else {
					// Connection to plugin lost, exit as it won't return
					log.Println("Torrent plugin has stopped in TorrentSync")
					time.Sleep(10 * time.Second)
					//return fmt.Errorf("Torrent plugin stopped")
				}
			}
		}

		// This tells our plugin to ignore any heights below this for retrieval
		s.DBStateManager.CompletedHeightTo(s.EntryDBHeightComplete)
		time.Sleep(rightDuration)
	}
}
