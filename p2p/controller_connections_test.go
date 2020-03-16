package p2p

import (
	"bytes"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func _detectProtocol(t *testing.T, version uint16) {
	conf := DefaultP2PConfiguration()
	conf.Network = NewNetworkID("test")
	conf.ProtocolVersion = version
	conf.NodeID = 1
	conf.ListenPort = "8999"

	A, B := net.Pipe()
	A.SetDeadline(time.Now().Add(time.Millisecond * 50))
	B.SetDeadline(time.Now().Add(time.Millisecond * 50))

	c := new(controller)
	c.net = new(Network)
	c.net.conf = &conf
	c.net.instanceID = 666

	sendprot := c.selectProtocol(B)
	sendshake := newHandshake(&conf, 123)

	go func() {
		if err := sendprot.SendHandshake(sendshake); err != nil {
			t.Error(err)
		}

		testp := newParcel(TypeMessage, []byte("foo"))
		if err := sendprot.Send(testp); err != nil {
			t.Error(err)
		}
	}()

	prot, hs, err := c.detectProtocolFromFirstMessage(A)
	if err != nil {
		t.Error(err)
	}

	if prot.Version() != version {
		t.Errorf("version mismatch. want = %d, got = %s", version, prot)
	}

	if !reflect.DeepEqual(sendshake, hs) {
		t.Errorf("handshake differs. want = %+v, got = %+v", sendshake, hs)
	}

	p, err := prot.Receive()
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(p.Payload, []byte("foo")) {
		t.Errorf("parcel didn't decode properly. want = %x, got = %x", []byte("foo"), p.Payload)
	}
}

func Test_controller_detectProtocol(t *testing.T) {
	_detectProtocol(t, 9)
	_detectProtocol(t, 10)
	_detectProtocol(t, 11)
}

func handshakeTestInstance(version uint16) *controller {
	conf := DefaultP2PConfiguration()
	conf.HandshakeTimeout = time.Second
	conf.ProtocolVersion = version
	conf.MaxIncoming = 2
	conf.ListenPort = "123"
	conf.ReadDeadline = time.Second
	conf.WriteDeadline = time.Second
	c := new(controller)
	c.peers = new(PeerStore)
	c.special = make(map[string]bool)
	c.net = new(Network)
	c.net.controller = c
	c.net.conf = &conf

	c.peerStatus = make(chan peerStatus, 2)
	c.peerData = make(chan peerParcel, 2)
	c.logger = log.WithField("package", "unit-test")
	c.net.instanceID = rand.Uint64()
	return c
}

func testControllerHandshakes(t *testing.T, name string, versionA, versionB, shouldVersion uint16) {
	// B connects to A
	A, B := net.Pipe()
	epA, _ := NewEndpoint("unittestincoming", "123") // B already has a valid port for epA
	epB, _ := NewEndpoint("unittestoutgoing", "1")   // A only knows the local port, not B's

	controllerA := handshakeTestInstance(versionA)
	controllerB := handshakeTestInstance(versionB)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		if _, _, err := controllerB.handshakeOutgoing(B, epA); err != nil { // instance B is dialing to ep A
			t.Errorf("[outgoing %s] handshake error: %v", name, err)
		}
		wg.Done()
	}()

	go func() {
		if err := controllerA.handshakeIncoming(A, epB); err != nil { // instance A receives incomplete ep B
			t.Errorf("[incoming %s] handshake error: %v", name, err)
		}
		wg.Done()
	}()

	wg.Wait()

	select {
	case inc := <-controllerA.peerStatus:
		if !inc.online {
			t.Errorf("[incoming %s] received offline signal", name)
		}
		p := inc.peer
		if !p.IsIncoming {
			t.Errorf("[incoming %s] peer not marked as incoming", name)
		}
		if p.prot.Version() != shouldVersion {
			t.Errorf("[incoming %s] invalid version. got = %d, want = %d", name, p.prot.Version(), shouldVersion)
		}

		if p.Endpoint.IP != "unittestoutgoing" {
			t.Errorf("[incoming %s] peer endpoint ip wrong. got = %s, want = %s", name, p.Endpoint.IP, "unittestoutgoing")
		}
		if p.Endpoint.Port != "123" {
			t.Errorf("[incoming %s] peer endpoint port wrong. got = %s, want = %s", name, p.Endpoint.Port, "123")
		}

		p.Stop()

	default:
		t.Errorf("[incoming %s] no peer status signal arrived for incoming", name)
	}

	select {
	case inc := <-controllerB.peerStatus:
		if !inc.online {
			t.Errorf("[outgoing %s] received offline signal", name)
		}
		p := inc.peer
		if p.IsIncoming {
			t.Errorf("[outgoing %s] peer marked as incoming", name)
		}
		if p.prot.Version() != shouldVersion {
			t.Errorf("[outgoing %s] invalid version. got = %d, want = %d", name, p.prot.Version(), shouldVersion)
		}

		if p.Endpoint.IP != "unittestincoming" {
			t.Errorf("[outgoing %s] peer endpoint ip wrong. got = %s, want = %s", name, p.Endpoint.IP, "unittestincoming")
		}
		if p.Endpoint.Port != "123" {
			t.Errorf("[outgoing %s] peer endpoint port wrong. got = %s, want = %s", name, p.Endpoint.Port, "123")
		}

		p.Stop()
	default:
		t.Errorf("[outgoing %s] no peer status signal arrived", name)
	}

}

func Test_controller_handshakes(t *testing.T) {
	testControllerHandshakes(t, "same version 9", 9, 9, 9)
	testControllerHandshakes(t, "same version 10", 10, 10, 10)
	testControllerHandshakes(t, "same version 11", 11, 11, 11)
	testControllerHandshakes(t, "backward compatible p2p1", 9, 10, 9)
	testControllerHandshakes(t, "agree on upper 9->11", 9, 11, 11)
	testControllerHandshakes(t, "agree on upper 10->11", 10, 11, 11)
	testControllerHandshakes(t, "agree on lower 10->9", 10, 9, 9)
	testControllerHandshakes(t, "agree on lower 11->9", 11, 9, 9)

}

func Test_controller_manageOnline(t *testing.T) {
	net := testNetworkHarness(t)

	done := make(chan bool, 1)
	go func() {
		net.controller.manageOnline()
		done <- true
	}()

	p := testRandomPeer(net)
	c := net.controller

	if c.peers.Total() != 0 {
		t.Fatalf("peerstore not empty, has %d peers", c.peers.Total())
	}
	c.peerStatus <- peerStatus{peer: p, online: true}
	time.Sleep(time.Millisecond * 50)

	if c.peers.Get(p.Hash) == nil {
		t.Errorf("peerstore get %s returned nil", p.Hash)
	}

	c.peerStatus <- peerStatus{peer: p, online: false}
	time.Sleep(time.Millisecond * 50)

	if p2 := c.peers.Get(p.Hash); p2 != nil {
		t.Errorf("peerstore peer %s was not removed: returned %s", p.Hash, p2.Hash)
	}

	// a second peer connecting with the same hash
	p2 := testRandomPeer(net)
	p2.Hash = p.Hash

	p.stopper.Do(func() {})

	c.peerStatus <- peerStatus{peer: p, online: true}
	c.peerStatus <- peerStatus{peer: p2, online: true}
	time.Sleep(time.Millisecond * 50)

	p3 := c.peers.Get(p.Hash)
	if p3 == nil {
		t.Errorf("duplicate peer not found")
	} else if p3 == p {
		t.Errorf("duplicate peer did not replace original peer")
	}

	close(net.stopper)
	<-done
}

// checks that the listen routine opens a port we can connect to via tcp
// also that stopping the network stops the listener
// it will fail the handshake but that's tested elsewhere
func Test_controller_listen(t *testing.T) {
	n1 := testNetworkHarness(t)
	ep, _ := NewEndpoint("127.0.0.1", "14236")
	n1.conf.BindIP = ep.IP
	n1.conf.ListenPort = ep.Port

	done := make(chan bool, 1)
	go func() {
		go n1.controller.listen()
		done <- true
	}()

	con, err := net.Dial("tcp", ep.String())
	if err != nil {
		t.Error(err)
	}
	con.Close()
	n1.Stop()
	<-done
}

// only checks that controller.Dial will open a tcp connection
func Test_controller_Dial(t *testing.T) {
	n1 := testNetworkHarness(t)

	ep, _ := NewEndpoint("127.0.0.1", "14237")
	listener, err := net.Listen("tcp", ep.String())
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan bool)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
		} else {
			conn.Close()
		}
		listener.Close()
		done <- true
	}()

	n1.controller.Dial(ep)

	<-done
}

func Test_controller_allowIncoming(t *testing.T) {
	net := testNetworkHarness(t)

	net.controller.setSpecial("special:1")
	banned := testRandomPeer(net)
	banned.stopper.Do(func() {})
	net.controller.peers.Add(banned)
	net.controller.ban(banned.Hash, time.Hour)

	sameAddr := Endpoint{IP: "samehost", Port: "1"}
	net.conf.PeerIPLimitIncoming = 1
	net.conf.MaxIncoming = 3

	type args struct {
		addr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ok", args{testRandomEndpoint().IP}, false},
		{"banned", args{banned.Endpoint.IP}, true},
		{"same ip limit", args{sameAddr.IP}, true},
		{"max peers", args{"max"}, true},
		{"special through max", args{"special"}, false},
	}
	for i, tt := range tests {
		if i == 2 { // same ip limit
			p := testRandomPeer(net)
			p.Endpoint = sameAddr
			net.controller.peers.Add(p)
		}
		if i == 3 { // max inc
			net.controller.peers.Add(testRandomPeer(net))
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := net.controller.allowIncoming(tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("controller.allowIncoming() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_controller_RejectWithShare(t *testing.T) {
	n := testNetworkHarness(t)
	for _, i := range []uint16{9, 11} {
		share := testRandomEndpointList(int(n.conf.PeerShareAmount))
		n.conf.ProtocolVersion = i

		A, B := net.Pipe()
		A.SetWriteDeadline(time.Now().Add(time.Millisecond * 100))
		B.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		go n.controller.RejectWithShare(A, share)

		prot := n.controller.selectProtocol(B)
		hs, err := prot.ReadHandshake()
		if err != nil {
			t.Errorf("prot %d receive error %v", i, err)
			continue
		}

		if hs.Type != TypeRejectAlternative {
			t.Errorf("prot %d parcel unexpected type. got = %s, want = %s", i, hs.Type, TypeRejectAlternative)
			continue
		}

		if !testEqualEndpointList(share, hs.Alternatives) {
			t.Errorf("prot %d different endpoint list. got = %v, want = %v", i, hs.Alternatives, share)
		}
	}
}
