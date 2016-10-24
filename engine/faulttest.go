// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"time"

	"math/rand"
	"os"
)

func waitToKill(k *bool) {
	t := rand.Int() % 120
	for t > 0 {
		os.Stderr.WriteString(fmt.Sprintf("     Will kill some servers in about %d seconds\n", t))
		t -= 10
		time.Sleep(10 * time.Second)
	}
	*k = true
}

// Wait some random amount of time between 0 and 2 minutes, and bring the node back.  We might
// come back before we are faulted, or we might not.
func bringback(f *FactomNode) {
	t := rand.Int() % 240
	for t > 0 {
		if !f.State.GetNetStateOff() {
			return
		}
		os.Stderr.WriteString(fmt.Sprintf("  Bringing %s back in %d seconds.\n", f.State.FactomNodeName, t))
		time.Sleep(10 * time.Second)
		t -= 10
	}
	f.State.SetNetStateOff(false) // Bring this node back
}

func offlineReport(faulting *bool) {
	for *faulting {
		// How many nodes are running.
		stmt := "Offline: "
		for _, f := range fnodes {
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
		var leaders []*FactomNode
		var running []*FactomNode

		lastgood := goodleaders
		goodleaders = 0
		// How many of the running nodes are leaders
		for _, f := range fnodes {
			if !f.State.GetNetStateOff() {
				running = append(running, f)
				if f.State.Leader {
					goodleaders++
				}
			}
			pl := f.State.ProcessLists.Get(f.State.LLeaderHeight)
			if pl != nil && len(pl.FedServers) > numleaders {
				numleaders = len(pl.FedServers)
			}
		}

		if lastgood != goodleaders {
			os.Stderr.WriteString(fmt.Sprintf("Of %d Leaders, we now have %d in working order.\n", numleaders, goodleaders))
		}

		for _, f := range running {
			if f.State.Leader {
				leaders = append(leaders, f)
			}
		}

		nextblk := false

		lastdbht := currentdbht
		lastminute := currentminute
		// Look at their process lists.  How many leaders do we expect?  What is the dbheight?
		for _, f := range fnodes {
			if int(f.State.LLeaderHeight) > currentdbht {
				currentminute = 0
				currentdbht = int(f.State.LLeaderHeight)
				nextblk = true
			}
			if !nextblk && f.State.CurrentMinute > currentminute {
				currentminute = f.State.CurrentMinute
			}
		}

		if !killing && goodleaders == numleaders {
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

		if killsome {
			killing = false
			// Wait some random amount of time.
			delta := rand.Int() % 20
			time.Sleep(time.Duration(delta) * time.Second)

			killsome = false
			kill := rand.Int() % ((numleaders / 2) - 2)
			kill++
			os.Stderr.WriteString(fmt.Sprintf("Killing %d of %d Leaders\n", kill, numleaders))
			for i := 0; i < kill; {
				n := rand.Int() % len(leaders)
				if !leaders[n].State.GetNetStateOff() {
					fmt.Sprintf(">>>> Killing %s", leaders[n].State.FactomNodeName)
					leaders[n].State.SetNetStateOff(true)
					go bringback(leaders[n])
					i++
				}
			}

			totalServerFaults += kill

		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
