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
	testControllerHandshakes(t, "agree on upper 9->10", 9, 10, 10)
	testControllerHandshakes(t, "agree on upper 9->11", 9, 11, 11)
	testControllerHandshakes(t, "agree on upper 10->11", 10, 11, 11)
	testControllerHandshakes(t, "agree on lower 10->9", 10, 9, 9)
	testControllerHandshakes(t, "agree on lower 11->9", 11, 9, 9)

}
