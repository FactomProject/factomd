package p2p

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func Test_controller_processPeerShare(t *testing.T) {
	net := testNetworkHarness(t)

	share := testRandomEndpointList(3)

	v9p := new(Peer)
	v9p.prot = newProtocolV9(net.conf.Network, net.conf.NodeID, net.conf.ListenPort, nil, nil)

	v10p := new(Peer)
	v10p.prot = newProtocolV10(nil, nil)

	v11p := new(Peer)
	v11p.prot = newProtocolV11(nil)

	tests := []struct {
		name string
		peer *Peer
		want []Endpoint
	}{
		{"v9", v9p, share},
		{"v10", v10p, share},
		{"v11", v11p, share},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := tt.peer.prot.MakePeerShare(share)
			if err != nil {
				t.Errorf("prot.MakePeerShare() = %v", err)
			}

			if len(payload) == 0 {
				t.Fatal("can't proceed with empty payload")
			}

			parcel := newParcel(TypePeerResponse, payload)

			if got := net.controller.processPeerShare(tt.peer, parcel); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("controller.processPeerShare() = %v, want %v", got, tt.want)
			}

			payload[0] += 1
			parcelBad := newParcel(TypePeerResponse, payload)

			if got := net.controller.processPeerShare(tt.peer, parcelBad); reflect.DeepEqual(got, tt.want) {
				t.Errorf("controller.processPeerShare() equal even though parcel changed")
			}
		})
	}
}

func Test_controller_shuffleTrimShare(t *testing.T) {
	net := testNetworkHarness(t)

	eps := testRandomEndpointList(64)
	net.conf.PeerShareAmount = 64

	got := net.controller.shuffleTrimShare(eps)

	if uint(len(got)) != net.conf.PeerShareAmount {
		t.Errorf("shuffleTrimShare() trimmed when it shouldn't have. len = %d, want = %d", len(got), net.conf.PeerShareAmount)
	} else {
		eq := true
		for i, ep := range eps {
			if !got[i].Equal(ep) {
				eq = false
				break
			}
		}
		if eq {
			t.Errorf("shuffleTrimShare() did not shuffle. order matched")
		}
	}

	net.conf.PeerShareAmount = 32

	got = net.controller.shuffleTrimShare(eps)
	if uint(len(got)) != net.conf.PeerShareAmount {
		t.Errorf("shuffleTrimShare() did not trim. len = %d, want = %d", len(got), net.conf.PeerShareAmount)
	}

}

func Test_controller_sharePeers(t *testing.T) {
	net := testNetworkHarness(t)
	p := testRandomPeer(net)
	p._setProtocol(10, nil) // only need to encode peer share, not send on wire

	share := testRandomEndpointList(128)
	net.controller.sharePeers(p, share)

	if len(p.send) == 0 {
		t.Error("peer did not receive a parcel")
	} else {
		parc := <-p.send

		list, err := p.prot.ParsePeerShare(parc.Payload)
		if err != nil {
			t.Fatalf("error unmarshalling peer share %v", err)
		}

		if !testEqualEndpointList(share, list) {
			t.Errorf("endpoint lists did not match")
		}
	}
}

func Test_controller_asyncPeerRequest(t *testing.T) {
	net := testNetworkHarness(t)
	peer := testRandomPeer(net)

	peer._setProtocol(10, nil) // no connection needed

	sendShare := testRandomEndpointList(int(net.conf.PeerShareAmount))

	done := make(chan bool)
	go func() {
		share, err := net.controller.asyncPeerRequest(peer)
		if err != nil {
			t.Errorf("async error %v", err)
		} else if !testEqualEndpointList(share, sendShare) {
			t.Errorf("async received different share")
		}
		done <- true
	}()

	time.Sleep(time.Millisecond * 100)

	net.controller.shareMtx.RLock()
	async, ok := net.controller.shareListener[peer.Hash]
	net.controller.shareMtx.RUnlock()

	if !ok {
		t.Fatal("no async channel found for peer")
	}

	if len(peer.send) == 0 {
		t.Errorf("peer did not send a request")
	} else if parc := <-peer.send; parc.ptype != TypePeerRequest {
		t.Errorf("parcel sent was not a request for peers. got = %s", parc.ptype)
	}

	payload, _ := peer.prot.MakePeerShare(sendShare)

	parc := newParcel(TypePeerResponse, payload)
	async <- parc

	<-done
}

func Test_controller_asyncPeerRequest_timeout(t *testing.T) {
	net := testNetworkHarness(t)
	net.conf.PeerShareTimeout = time.Millisecond * 50
	peer := testRandomPeer(net)

	got, err := net.controller.asyncPeerRequest(peer)
	if err == nil {
		t.Errorf("async did not return a timeout error. got = %v", got)
	}
}

// testing this is challenging but there are some expectations we can have
// 1. the bootstrap peers will be dialed first
// 2. the special peers will be dialed
// 3. the seed peers will be dialed
// and the replenish loop will terminate when the network stops
func Test_controller_catReplenish(t *testing.T) {
	//dialed := make(map[Endpoint]bool)

	testList := func(name string) []Endpoint {
		list := make([]Endpoint, 3)
		for i := range list {
			list[i] = Endpoint{IP: name, Port: fmt.Sprintf("%d", i+1)}
		}
		return list
	}

	net := testNetworkHarness(t)
	// create new dialer with lower timeout
	net.conf.DialTimeout = time.Millisecond
	net.conf.RedialInterval = time.Minute // we want to exhaust each endpoint in this unit test
	net.controller.dialer, _ = NewDialer("", net.conf.RedialInterval, net.conf.DialTimeout)

	net.controller.bootstrap = testList("bootstrap")

	net.controller.seed.cache = testList("seed")
	net.controller.seed.cacheTTL = time.Hour
	net.controller.seed.cacheTime = time.Now()

	net.controller.specialEndpoints = testList("special")

	ap := testRandomPeer(net)
	ap._setEndpoint(Endpoint{IP: "async", Port: "1"})
	ap._setProtocol(10, nil)
	net.controller.peers.Add(ap)
	asyncShare := testList("share")

	done := make(chan bool)
	go func() {
		net.controller.catReplenish()
		done <- true
	}()

	go func() {
		parc := <-ap.send
		if parc.ptype != TypePeerRequest {
			t.Errorf("async peer did not receive right parcel: %s", parc.ptype)
		}

		async := net.controller.shareListener[ap.Hash]

		pl, _ := ap.prot.MakePeerShare(asyncShare)
		resp := newParcel(TypePeerResponse, pl)
		async <- resp
	}()

	time.Sleep(time.Millisecond * 100)
	net.Stop()
	<-done

	for i, ep := range net.controller.bootstrap {
		if net.controller.dialer.CanDial(ep) {
			t.Errorf("bootstrap - did not dial ep %d", i)
		}
	}

	for i, ep := range net.controller.seed.retrieve() {
		if net.controller.dialer.CanDial(ep) {
			t.Errorf("seed- did not dial ep %d", i)
		}
	}

	for i, ep := range net.controller.specialEndpoints {
		if net.controller.dialer.CanDial(ep) {
			t.Errorf("special - did not dial ep %d", i)
		}
	}

	any := false
	for _, ep := range asyncShare {
		if !net.controller.dialer.CanDial(ep) {
			any = true
		}
	}

	if !any {
		t.Errorf("share - did not dial any of the shared endpoints")
	}

}

func Test_controller_makePeerShare(t *testing.T) {
	net := testNetworkHarness(t)
	list := testRandomEndpointList(int(net.conf.PeerShareAmount + 1))
	for _, ep := range list {
		p := testRandomPeer(net)
		p._setEndpoint(ep)
		net.controller.peers.Add(p)
	}

	for _, ep := range list {
		share := net.controller.makePeerShare(ep)
		if len(share) != len(list)-1 {
			t.Errorf("peer share without %s yielded wrong count. got = %d, want = %d", ep, len(share), len(list)-1)
		} else {
			for _, in := range share {
				if in.Equal(ep) {
					t.Errorf("peer share without %s included %s", ep, in)
				}
			}
		}
	}
}

func Test_controller_runCatRound(t *testing.T) {
	n := testNetworkHarness(t)

	rounds := n.Rounds()

	n.conf.DropTo = 1

	// create 3 peers, 2 should be dropped
	for i := 0; i < 3; i++ {
		rp := testRandomPeer(n)
		A, B := net.Pipe()
		rp.conn = A
		defer B.Close()
		n.controller.peers.Add(rp)
	}

	n.controller.lastRound = time.Time{}
	n.controller.runCatRound()

	if rounds == n.Rounds() {
		t.Errorf("net.Rounds() did not increase. got = %d, want = %d", n.Rounds(), rounds+1)
	}

	if len(n.controller.peerStatus) != 2 {
		t.Errorf("not enough status messages in peerStatus. got = %d, want = 2", len(n.controller.peerStatus))

		for len(n.controller.peerStatus) > 0 {
			if (<-n.controller.peerStatus).online {
				t.Errorf("peer went online instead of offline")
			}
		}
	}
}
