// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
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

func initializeEmptyDiscovery() (*DTE, Discovery) {
	dte := startDiscoveryTestEnvironment()
	d := Discovery{}
	d.logger = discoLogger
	d.knownPeers = map[string]Peer{}
	d.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	d.peersFilePath = dte.peersFile.Name()
	d.seedURL = dte.goodSeedUrl()
	return dte, d
}

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
	dte := startDiscoveryTestEnvironment()
	defer dte.cleanup()

	discovery := new(Discovery).Init(dte.peersFile.Name(), dte.goodSeedUrl())

	seeds := dte.createSeedPeers()
	valids := dte.createValidPeers()
	outdated := dte.createOutdatedPeers()
	expected := len(seeds) + len(valids) + len(outdated)

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

func TestDiscovery_isPeerPresent(t *testing.T) {
	dte, disco := initializeEmptyDiscovery()
	defer dte.cleanup()
	peers := dte.createValidPeers()

	for _, p := range peers {
		if disco.isPeerPresent(*p) {
			t.Errorf("Peer %s is present when it shouldn't be", p.AddressPort())
		}

		disco.updatePeer(*p)

		if !disco.isPeerPresent(*p) {
			t.Errorf("Peer %s is not present when it should be", p.AddressPort())
		}
	}
}

func TestDiscovery_updatePeer(t *testing.T) {
	dte := startDiscoveryTestEnvironment()
	defer dte.cleanup()
	discovery := new(Discovery).Init(dte.peersFile.Name(), dte.goodSeedUrl())

	peers := dte.createValidPeers()

	if _, ok := discovery.getPeer("invalid address"); ok {
		t.Errorf("Found non-existent peer")
	}

	for _, p := range peers {
		pOld, ok := discovery.getPeer(p.Address)

		if !ok {
			t.Fatal("Discovery failed to initialize properly")
		}

		if pOld.Address != p.Address || pOld.Port != p.Port {
			t.Fatal("Discovery returned the wrong peer")
		}

		p.QualityScore = 200
		discovery.updatePeer(*p)

		pNew, ok := discovery.getPeer(p.Address)

		if !ok {
			t.Fatalf("Peer %s disappeared after updating", p.AddressPort())
		}

		if pNew.QualityScore != 200 {
			t.Errorf("Peer %s failed to update qualityscore (should be 200, is %d)", pNew.AddressPort(), pNew.QualityScore)
		}
	}
}

func TestDiscovery_getPeer(t *testing.T) {
	dte, disco := initializeEmptyDiscovery()
	defer dte.cleanup()
	peers := dte.createValidPeers()

	for _, p := range peers {
		if disco.isPeerPresent(*p) {
			t.Errorf("Peer %s is present when it shouldn't be", p.AddressPort())
		}
		disco.updatePeer(*p)

		pNew, exists := disco.getPeer(p.Address)

		if !exists || pNew.Address != p.Address || pNew.Port != p.Port {
			t.Errorf("Peer %s is not present when it should be", p.AddressPort())
		}
	}
}

func TestDiscovery_PeerCount(t *testing.T) {
	dte := startDiscoveryTestEnvironment()
	defer dte.cleanup()

	discovery := new(Discovery).Init(dte.peersFile.Name(), dte.goodSeedUrl())

	seeds := dte.createSeedPeers()
	valids := dte.createValidPeers()
	outdated := dte.createOutdatedPeers()
	expected := len(seeds) + len(valids) + len(outdated)

	if expected != discovery.PeerCount() {
		t.Errorf("Expected %d peers, only have %d", expected, discovery.PeerCount())
	}
}

func TestDiscovery_LoadPeers(t *testing.T) {
	dte, disco := initializeEmptyDiscovery()
	defer dte.cleanup()

	if disco.PeerCount() != 0 {
		t.Fatalf("Test setup did not initialize properly")
	}

	// attempt loading from a blank file
	blank, err := ioutil.TempFile(os.TempDir(), "peers-*.json")
	if err != nil {
		log.Fatalln(err)
	}

	disco.peersFilePath = blank.Name()
	disco.LoadPeers()
	if disco.PeerCount() != 0 {
		t.Fatalf("Discovery generated peers from an empty file")
	}
	os.Remove(blank.Name())

	// load from test environment
	disco.peersFilePath = dte.peersFile.Name()
	disco.LoadPeers()

	valids := dte.createValidPeers()
	outdated := dte.createOutdatedPeers()
	expected := len(valids) + len(outdated)

	if l := disco.PeerCount(); l != expected {
		t.Errorf("%d peers found, %d expected", l, expected)
	}

	for _, p := range valids {
		if !disco.isPeerPresent(*p) {
			t.Errorf("Valids peer %s not present in discovery", p.AddressPort())
		}
	}
	for _, p := range outdated {
		if !disco.isPeerPresent(*p) {
			t.Errorf("Outdated peer %s not present in discovery", p.AddressPort())
		}
	}
}

func TestDiscovery_SavePeers(t *testing.T) {
	dte, disco := initializeEmptyDiscovery()
	defer dte.cleanup()

	peers := []*Peer{
		dte.createPeer("1.0.0.1", "1234", 0, RegularPeer),                            // good
		dte.createPeer("1.0.0.2", "1234", 0, SpecialPeerCmdLine),                     // special
		dte.createPeer("1.0.0.3", "1234", MinumumQualityScore-1, SpecialPeerCmdLine), // special but bad
		dte.createPeer("1.0.0.4", "1234", MinumumQualityScore-1, RegularPeer),        // bad
		dte.createPeer("1.0.0.5", "1234", 200, RegularPeer),                          // good but old
	}
	peers[4].LastContact = time.Now().AddDate(-1, 0, 0) // make old
	peerok := []bool{true, true, true, false, false}

	for _, p := range peers { // add
		disco.updatePeer(*p)
	}

	if disco.PeerCount() != len(peers) {
		t.Fatalf("Test environment did not set up correctly")
	}

	disco.SavePeers()
	disco.knownPeers = make(map[string]Peer) // reset
	disco.LoadPeers()

	for i, p := range peers {
		if _, ok := disco.getPeer(p.Address); ok != peerok[i] {
			if peerok[i] {
				t.Errorf("Good peer %s not saved to file", p.AddressPort())
			} else {
				t.Errorf("Bad peer %s was saved to file but shouldn't have been", p.AddressPort())
			}
		}
	}
}
