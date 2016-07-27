package controlPanel_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/p2p"
)

var _ = fmt.Sprintf("")

func TestFormatDuration(t *testing.T) {
	initial := time.Now().Add(-49 * time.Hour)
	if FormatDuration(initial) != "2 days" {
		t.Errorf("Time display incorrect : days")
	}

	initial = time.Now().Add(-25 * time.Hour)
	if FormatDuration(initial) != "1 day" {
		t.Errorf("Time display incorrect : days")
	}

	initial = time.Now().Add(-23 * time.Hour)
	if FormatDuration(initial) != "23 hrs" {
		t.Errorf("Time display incorrect : hrs")
	}

	initial = time.Now().Add(-1 * time.Hour)
	if FormatDuration(initial) != "1 hr" {
		t.Errorf("Time display incorrect : hr")
	}

	initial = time.Now().Add(-59 * time.Minute)
	if FormatDuration(initial) != "59 mins" {
		t.Errorf("Time display incorrect : mins")
	}

	initial = time.Now().Add(-1 * time.Minute)
	if FormatDuration(initial) != "1 min" {
		t.Errorf("Time display incorrect : min")
	}

	initial = time.Now().Add(-30 * time.Second)
	if FormatDuration(initial) != "30 secs" {
		t.Errorf("Time display incorrect : secs")
	}
}

func TestTallyTotals(t *testing.T) {
	cm := NewConnectionsMap()
	var i uint32
	for i = 0; i < 10; i++ {
		cm.Connect(fmt.Sprintf("%d", i), NewP2PConnection(i, i, i, i, fmt.Sprintf("%d", i), i))
	}
	for i = 10; i < 20; i++ {
		cm.Disconnect(fmt.Sprintf("%d", i), NewP2PConnection(i, i, i, i, fmt.Sprintf("%d", i), i))
	}
	cm.TallyTotals()
	if cm.Totals.BytesSentTotal != 190 {
		t.Errorf("Byte Sent does not match")
	}
	if cm.Totals.BytesReceivedTotal != 190 {
		t.Errorf("Byte Received does not match")
	}
	if cm.Totals.MessagesSent != 190 {
		t.Errorf("Msg Sent does not match")
	}
	if cm.Totals.MessagesReceived != 190 {
		t.Errorf("Msg Received does not match")
	}
	if cm.Totals.PeerQualityAvg != 4 {
		t.Errorf("Peer Quality does not match %d", cm.Totals.PeerQualityAvg)
	}

	for key := range cm.GetConnectedCopy() {
		cm.RemoveConnection(key)
	}
	for key := range cm.GetDisconnectedCopy() {
		cm.RemoveConnection(key)
	}
	cm.TallyTotals()
	if cm.Totals.BytesSentTotal != 0 {
		t.Errorf("Byte Sent does not match")
	}
	if cm.Totals.BytesReceivedTotal != 0 {
		t.Errorf("Byte Received does not match")
	}
	if cm.Totals.MessagesSent != 0 {
		t.Errorf("Msg Sent does not match")
	}
	if cm.Totals.MessagesReceived != 0 {
		t.Errorf("Msg Received does not match")
	}
	if cm.Totals.PeerQualityAvg != 0 {
		t.Errorf("Peer Quality does not match %d", cm.Totals.PeerQualityAvg)
	}
}

// Absurd map accessing
func TestConcurrency(t *testing.T) {
	cm := NewConnectionsMap()
	var i uint32
	for i = 0; i < 100; i++ {
		// Random Connections
		go func() {
			randPeers := make([]p2p.ConnectionMetrics, 0)
			for ii := 0; ii < 10; ii++ {
				randPeer := NewRandomP2PConnection()
				cm.Connect(randPeer.PeerAddress, randPeer)
				cm.TallyTotals()

				randPeer2 := NewRandomP2PConnection()
				cm.AddConnection(randPeer2.PeerAddress, *randPeer2)

				randPeers = append(randPeers, *randPeer)
				randPeers = append(randPeers, *randPeer2)
			}
			for _, peer := range randPeers {
				cm.TallyTotals()
				cm.Disconnect(peer.PeerAddress, cm.GetConnection(peer.PeerAddress))
			}
			cm.CleanDisconnected()
		}()

		go func() {
			randPeers := make([]p2p.ConnectionMetrics, 0)
			for ii := 0; ii < 10; ii++ {
				randPeer1 := NewSeededP2PConnection(i)
				cm.Connect(randPeer1.PeerAddress, randPeer1)
				cm.TallyTotals()

				randPeer2 := NewSeededP2PConnection(i)
				cm.AddConnection(randPeer2.PeerAddress, *randPeer2)

				randPeers = append(randPeers, *randPeer1)
				randPeers = append(randPeers, *randPeer2)
			}
			for _, peer := range randPeers {
				cm.TallyTotals()
				cm.Disconnect(peer.PeerAddress, cm.GetConnection(peer.PeerAddress))
			}
			cm.CleanDisconnected()
		}()
		// Sharing connections
	}
}

func NewRandomP2PConnection() *p2p.ConnectionMetrics {
	con := NewP2PConnection(rand.Uint32(), rand.Uint32(), rand.Uint32(), rand.Uint32(), string(rand.Uint32()), rand.Uint32())
	return con
}

func NewSeededP2PConnection(seed uint32) *p2p.ConnectionMetrics {
	con := NewP2PConnection(seed, seed, seed, seed, string(seed), seed)
	return con
}

func NewP2PConnection(bs uint32, br uint32, ms uint32, mr uint32, addr string, pq uint32) *p2p.ConnectionMetrics {
	pc := new(p2p.ConnectionMetrics)
	pc.MomentConnected = time.Now()
	pc.BytesSent = bs
	pc.BytesReceived = br
	pc.MessagesSent = ms
	pc.MessagesReceived = mr
	pc.PeerAddress = addr
	pc.PeerQuality = int32(pq)

	return pc
}
