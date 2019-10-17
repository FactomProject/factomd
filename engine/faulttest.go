// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"github.com/FactomProject/factomd/fnode"
	"time"

	"math/rand"
	"os"
)

func waitToKill(k *bool) {
	t := rand.Int()%120 + 60
	for t > 0 {
		os.Stderr.WriteString(fmt.Sprintf("     Will kill some servers in about %d seconds\n", t))
		if t < 30 {
			time.Sleep(time.Duration(t) * time.Second)
		} else {
			time.Sleep(30 * time.Second)
		}
		t -= 30
	}
	*k = true
}

// Wait some random amount of time between 0 and 2 minutes, and bring the node back.  We might
// come back before we are faulted, or we might not.
func bringback(f *fnode.FactomNode) {
	t := rand.Int()%120 + 60
	for t > 0 {
		if !f.State.GetNetStateOff() {
			return
		}
		os.Stderr.WriteString(fmt.Sprintf("  Bringing %s back in %d seconds.\n", f.State.FactomNodeName, t))
		if t < 30 {
			time.Sleep(time.Duration(t) * time.Second)
		} else {
			time.Sleep(30 * time.Second)
		}
		t -= 30
	}
	f.State.SetNetStateOff(false) // Bring this node back
}

func offlineReport(faulting *bool) {
	for *faulting {
		// How many nodes are running.
		stmt := "Offline: "
		for _, f := range fnode.GetFnodes() {
			if f.State.GetNetStateOff() {
				stmt = stmt + fmt.Sprintf(" %s", f.State.FactomNodeName)
			}
		}
		if len(stmt) > 10 {
			os.Stderr.WriteString(stmt + "\n")
		}

		time.Sleep(20 * time.Second)
	}

}

func faultTest(faulting *bool) {
	killsome := false
	killing := false
	numleaders := 0
	currentdbht := 0
	currentminute := 0
	goodleaders := 0

	go offlineReport(faulting)

	for *faulting {
		var leaders []*fnode.FactomNode

		lastgood := goodleaders
		goodleaders = 0
		// How many of the running nodes are leaders
		for _, f := range fnode.GetFnodes() {
			if f.State.GetNetStateOff() {
				continue
			}
			if !f.State.Leader {
				continue
			}
			if int(f.State.LLeaderHeight) < currentdbht {
				continue
			}
			if int(f.State.LLeaderHeight) == currentdbht && int(f.State.CurrentMinute) < currentminute {
				continue
			}

			goodleaders++
			leaders = append(leaders, f)

			pl := f.State.LeaderPL
			if pl != nil && len(pl.FedServers) > numleaders {
				numleaders = len(pl.FedServers)
			}
		}

		if lastgood != goodleaders {
			os.Stderr.WriteString(fmt.Sprintf("Of %d Leaders, we now have %d in working order.\n", numleaders, goodleaders))
		}

		nextblk := false

		lastdbht := currentdbht
		lastminute := currentminute
		// Look at their process lists.  How many leaders do we expect?  What is the dbheight?
		for _, f := range fnode.GetFnodes() {
			if int(f.State.LLeaderHeight) > currentdbht {
				currentminute = 0
				currentdbht = int(f.State.LLeaderHeight)
				nextblk = true
			}
			if !nextblk && f.State.CurrentMinute > currentminute {
				currentminute = f.State.CurrentMinute
			}
		}

		if !killing && goodleaders >= numleaders {
			if currentdbht > lastdbht || currentminute > lastminute {
				killing = true
				go waitToKill(&killsome)
			}
		}
		// Can't run this test without at least three leaders.
		if numleaders < 3 {
			os.Stderr.WriteString("Not enough leaders to run fault test\n")
			*faulting = false
			return
		}

		if killsome && len(leaders) > 0 && goodleaders >= numleaders {
			killing = false
			killsome = false
			// Wait some random amount of time.
			delta := rand.Int() % 20
			time.Sleep(time.Duration(delta) * time.Second)

			kill := 1
			maxLeadersToKill := numleaders / 2
			if maxLeadersToKill == 0 {
				maxLeadersToKill = 1
			} else {
				kill = rand.Int() % maxLeadersToKill
				kill++
			}
			kill = 1

			os.Stderr.WriteString(fmt.Sprintf("Killing %3d of %3d Leaders\n", kill, numleaders))
			for i := 0; i < kill; {
				n := rand.Int() % len(leaders)
				if !leaders[n].State.GetNetStateOff() {
					os.Stderr.WriteString(fmt.Sprintf("     >>>> Killing %10s %s\n",
						leaders[n].State.FactomNodeName,
						leaders[n].State.GetIdentityChainID().String()[4:16]))
					leaders[n].State.SetNetStateOff(true)
					go bringback(leaders[n])
					i++
					time.Sleep(time.Duration(rand.Int()%40) * time.Second)
					totalServerFaults++
				}
			}

		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
