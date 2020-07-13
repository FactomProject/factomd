package p2p

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func Test_controller_route(t *testing.T) {
	net := testNetworkHarness(t)
	net.conf.Fanout = 4
	net.conf.MaxPeers = 8
	for i := 0; i < int(net.conf.MaxPeers); i++ {
		net.controller.peers.Add(testRandomPeer(net))
	}

	done := make(chan bool)
	go func() {
		net.controller.route()
		done <- true
	}()

	rp := net.controller.randomPeer()

	single := testRandomParcel()
	single.Address = rp.Hash

	net.toNetwork <- single

	select {
	case arr := <-rp.send:
		if arr != single {
			t.Errorf("single target parcel changed. got = %v, want = %v", arr, single)
		}
	case <-time.After(time.Millisecond * 50):
		t.Errorf("single target parcel did not arrive")
	}

	rand := testRandomParcel()
	rand.Address = RandomPeer
	net.toNetwork <- rand

	time.Sleep(time.Millisecond * 50)

	parcels := 0
	for _, p := range net.controller.peers.Slice() {
		for len(p.send) > 0 {
			parc := <-p.send
			if parc != rand {
				t.Errorf("unexpected parcel %v showed up", parc)
			}
			parcels++
		}
	}

	if parcels != 1 {
		t.Errorf("received incorrect amount of parcels. got = %d, want = %d", parcels, 1)
	}

	broadcast := testRandomParcel()
	broadcast.Address = Broadcast
	net.toNetwork <- broadcast

	time.Sleep(time.Millisecond * 50)

	parcels = 0
	for _, p := range net.controller.peers.Slice() {
		pparcels := 0
		for len(p.send) > 0 {
			parc := <-p.send
			if parc != broadcast {
				t.Errorf("unexpected broadcast parcel %v showed up", parc)
			}
			pparcels++
		}
		if pparcels > 1 {
			t.Errorf("peer received multiple broadcast parcels. got = %d, want <= 1", pparcels)
		}
		parcels += pparcels
	}

	if parcels != int(net.conf.Fanout) {
		t.Errorf("received incorrect amount of broadcast parcels. got = %d, want = %d", parcels, net.conf.Fanout)
	}

	fullBroadcast := testRandomParcel()
	fullBroadcast.Address = FullBroadcast
	net.toNetwork <- fullBroadcast

	time.Sleep(time.Millisecond * 50)

	for _, p := range net.controller.peers.Slice() {
		if len(p.send) != 1 {
			t.Errorf("peer did not receive right amount of full broadcast parcels. got = %d, want = 1", len(p.send))
		} else {
			parc := <-p.send
			if parc != fullBroadcast {
				t.Errorf("unexpected full broadcast parcel %v showed up", parc)
			}
		}
	}

	net.Stop()
	<-done
}

func Test_controller_manageData(t *testing.T) {
	net := testNetworkHarness(t)

	net.controller.peerData <- peerParcel{testRandomPeer(net), testParcel(TypeAlert)}
	net.controller.peerData <- peerParcel{nil, testParcel(TypePing)} // gets tossed

	ping := testRandomPeer(net)
	net.controller.peerData <- peerParcel{ping, testParcel(TypePing)} // payload shouldn't matter

	share := testRandomPeer(net)
	share.lastPeerRequest = time.Time{}
	share._setProtocol(10, nil)
	net.controller.peerData <- peerParcel{share, testParcel(TypePeerRequest)}

	async := make(chan *Parcel, 1)
	asyncPeer := testRandomPeer(net)
	net.controller.shareListener[asyncPeer.Hash] = async // not thread safe assignment
	net.controller.peerData <- peerParcel{asyncPeer, testParcel(TypePeerResponse)}

	msg1 := testParcel(TypeMessage)
	net.controller.peerData <- peerParcel{nil, msg1}
	msg2 := testParcel(TypeMessagePart)
	net.controller.peerData <- peerParcel{nil, msg2}

	finish := newParcel(TypeMessage, []byte("finish"))
	net.controller.peerData <- peerParcel{nil, finish}
	go func() {

		next := func() *Parcel {
			select {
			case <-time.After(time.Second):
				t.Fatal("stuck waiting for application parcel")
			case p := <-net.fromNetwork:
				return p
			}
			return nil
		}

		if app1 := next(); app1 != msg1 {
			t.Errorf("msg1 did not send first")
		}
		if app2 := next(); app2 != msg2 {
			t.Errorf("msg2 did not send second")
		} else if app2.ptype != TypeMessage {
			t.Errorf("msgs2 did not transform. got = %s, want = %s", app2.ptype, TypeMessage)
		}

		f := next()
		if string(f.Payload) != "finish" {
			t.Errorf("finish app message did not arrive. got = %v", f)
		}
		net.Stop()
	}()

	net.controller.manageData()

	if len(ping.send) != 1 {
		t.Errorf("ping peer did not receive pong")
	} else if p := <-ping.send; p.ptype != TypePong {
		t.Errorf("ping peer did not send right response. got = %s, want = %s", p.ptype, TypePong)
	}

	if len(share.send) != 1 {
		t.Errorf("share peer did not receive share")
	} else if p := <-share.send; p.ptype != TypePeerResponse {
		t.Errorf("share peer did not send right response. got = %s, want = %s", p.ptype, TypePeerResponse)
	}

	if len(async) != 1 {
		t.Errorf("async response did not deliver")
	} else if p := <-async; p.ptype != TypePeerResponse {
		t.Errorf("peer response did not send right response. got = %s, want = %s", p.ptype, TypePeerResponse)
	}
}

func Test_controller_selectNoReplayPeers(t *testing.T) {
	payload := make([]byte, 1024)
	rand.Read(payload)
	hash := sha1.Sum(payload)

	net := testNetworkHarness(t)
	net.conf.MaxPeers = 8
	for i := 0; i < int(net.conf.MaxPeers); i++ {
		net.controller.peers.Add(testRandomPeer(net))
	}

	peers := net.controller.peers.Slice()

	for i := 0; i < 4; i++ {
		peers[i].resend.Add(hash)
		fmt.Printf("registered hash with peer %s\n", peers[i].Hash)
	}

	filtered := net.controller.selectNoResendPeers(payload)
	if len(filtered) != 4 {
		t.Errorf("invalid number of peers selected. got = %d, want = 4", len(filtered))
	}
	for _, f := range filtered {
		for i := 0; i < 4; i++ {
			if f.Hash == peers[i].Hash {
				t.Errorf("peer %s was mistakenly included", f.Hash)
			}
		}
	}
}
