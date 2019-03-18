// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

type DTE struct {
	address   string
	server    *http.Server
	peersFile *os.File
	time      time.Time
}

const (
	Network1     NetworkID = 0x1
	Network2     NetworkID = 0x2
	GoodPeerFile           = "{\"2.0.0.1:8108\":{\"QualityScore\":0,\"Address\":\"2.0.0.1\",\"Port\":\"8108\",\"NodeID\":0,\"Hash\":\"2.0.0.1:8108 380704bb7b4d7c03\",\"Location\":33554433,\"Network\":1,\"Type\":0,\"Connections\":0,\"LastContact\":\"%[1]s\",\"Source\":{}},\"2.0.0.2:8108\":{\"QualityScore\":0,\"Address\":\"2.0.0.2\",\"Port\":\"8108\",\"NodeID\":0,\"Hash\":\"2.0.0.2:8108 365a858149c6e2d1\",\"Location\":33554434,\"Network\":1,\"Type\":0,\"Connections\":0,\"LastContact\":\"%[1]s\",\"Source\":{}},\"2.0.0.3:8108\":{\"QualityScore\":0,\"Address\":\"2.0.0.3\",\"Port\":\"8108\",\"NodeID\":0,\"Hash\":\"2.0.0.3:8108 57e9d1860d1d68d8\",\"Location\":33554435,\"Network\":1,\"Type\":0,\"Connections\":0,\"LastContact\":\"%[1]s\",\"Source\":{}},\"3.0.0.1:8108\":{\"QualityScore\":0,\"Address\":\"3.0.0.1\",\"Port\":\"8108\",\"NodeID\":0,\"Hash\":\"3.0.0.1:8108 866cb397916001e\",\"Location\":50331649,\"Network\":1,\"Type\":0,\"Connections\":0,\"LastContact\":\"%[2]s\",\"Source\":{}},\"3.0.0.2:8108\":{\"QualityScore\":0,\"Address\":\"3.0.0.2\",\"Port\":\"8108\",\"NodeID\":0,\"Hash\":\"3.0.0.2:8108 1408d2ac22c4d294\",\"Location\":50331650,\"Network\":1,\"Type\":0,\"Connections\":0,\"LastContact\":\"%[2]s\",\"Source\":{}},\"3.0.0.3:8108\":{\"QualityScore\":0,\"Address\":\"3.0.0.3\",\"Port\":\"8108\",\"NodeID\":0,\"Hash\":\"3.0.0.3:8108 c697f48392907a0\",\"Location\":50331651,\"Network\":1,\"Type\":0,\"Connections\":0,\"LastContact\":\"%[2]s\",\"Source\":{}}}"
)

func startDiscoveryTestEnvironment() (dte *DTE) {
	dte = &DTE{address: "127.0.0.1:8081"}
	dte.time = time.Now()

	log.SetLevel(log.DebugLevel)
	mux := http.NewServeMux()
	mux.HandleFunc("/badseed", dte.badSeed)
	mux.HandleFunc("/goodseed", dte.goodSeed)

	dte.server = &http.Server{Addr: dte.address, Handler: mux}
	go dte.server.ListenAndServe()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "peers-*.json")
	if err != nil {
		log.Fatalln(err)
	}
	dte.peersFile = tmpFile
	writer := bufio.NewWriter(tmpFile)
	t1, _ := dte.time.MarshalText()
	t2, _ := dte.time.AddDate(-1, 0, 0).MarshalText()
	writer.WriteString(fmt.Sprintf(GoodPeerFile, t1, t2))
	writer.Flush()

	// set global defaults
	CurrentNetwork = Network1
	NumberPeersToConnect = 32

	return dte
}

func (dte *DTE) badSeed(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(rw, "2.2.2.2") // bad port
}

func (dte *DTE) goodSeed(rw http.ResponseWriter, req *http.Request) {
	for _, p := range dte.createSeedPeers() {
		fmt.Fprintln(rw, fmt.Sprintf("%s:%s", p.Address, p.Port))
	}
}

func (dte *DTE) cleanup() {
	dte.server.Shutdown(context.Background())
	os.Remove(dte.peersFile.Name())
}

func (dte *DTE) createPeer(address, port string, quality int32, peerType uint8) *Peer {
	p := Peer{}
	p.Init(address, port, quality, peerType, 0) // suppress errors
	p.LastContact = dte.time
	return &p
}

func (dte *DTE) goodSeedUrl() string {
	return fmt.Sprintf("http://%s/goodseed", dte.address)
}

func (dte *DTE) createSeedPeers() []*Peer {
	return []*Peer{
		dte.createPeer("1.0.0.1", "8108", 0, RegularPeer),
		dte.createPeer("1.0.0.2", "8108", 0, RegularPeer),
		dte.createPeer("1.0.0.3", "8108", 0, RegularPeer),
	}
}

func (dte *DTE) createValidPeers() []*Peer {
	return []*Peer{
		dte.createPeer("2.0.0.1", "8108", 0, RegularPeer),
		dte.createPeer("2.0.0.2", "8108", 0, RegularPeer),
		dte.createPeer("2.0.0.3", "8108", 0, RegularPeer),
	}
}

func (dte *DTE) createOutdatedPeers() []*Peer {
	p := []*Peer{
		dte.createPeer("3.0.0.1", "8108", 0, RegularPeer),
		dte.createPeer("3.0.0.2", "8108", 0, RegularPeer),
		dte.createPeer("3.0.0.3", "8108", 0, RegularPeer),
	}
	for i := range p {
		p[i].LastContact = dte.time.AddDate(-1, 0, 0)
	}
	return p
}

func TestDiscovery_Init(t *testing.T) {
	fmt.Println("Starting test environment...")
	dte := startDiscoveryTestEnvironment()
	defer dte.cleanup()

	discovery := new(Discovery).Init(dte.peersFile.Name(), dte.goodSeedUrl())

	seeds := dte.createSeedPeers()
	valids := dte.createValidPeers()
	outdated := dte.createOutdatedPeers()
	expected := len(seeds) + len(valids) + len(outdated)

	for i, p := range discovery.knownPeers {
		fmt.Println("peer", i, p)
	}

	if l := len(discovery.knownPeers); l != expected {
		t.Errorf("%d peers found, %d expected", l, expected)
	}

	for _, p := range seeds {
		if !discovery.isPeerPresent(*p) {
			t.Errorf("Seed peer %s not present in discovery", p.AddressPort())
		}
	}
	for _, p := range valids {
		if !discovery.isPeerPresent(*p) {
			t.Errorf("Valids peer %s not present in discovery", p.AddressPort())
		}
	}
	for _, p := range outdated {
		if !discovery.isPeerPresent(*p) {
			t.Errorf("Outdated peer %s not present in discovery", p.AddressPort())
		}
	}
}
