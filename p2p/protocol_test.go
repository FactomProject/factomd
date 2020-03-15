package p2p

import (
	"encoding/gob"
	"io"
	"math/rand"
	"net"
	"reflect"
	"testing"
	"time"
)

func testEqualEndpointList(a, b []Endpoint) bool {
	if len(a) != len(b) {
		return false
	}
	has := make(map[Endpoint]bool)
	for _, ep := range a {
		has[ep] = true
	}
	for _, ep := range b {
		if !has[ep] {
			return false
		}
	}
	return true
}

func testRandomParcel() *Parcel {
	return testParcel(ParcelType(rand.Intn(len(typeStrings))))
}
func testParcel(typ ParcelType) *Parcel {
	p := new(Parcel)
	p.ptype = typ
	p.Address = ""                              // address is internal in prot10+
	p.Payload = make([]byte, 1+rand.Intn(8191)) // 1 byte - 8KiB
	rand.Read(p.Payload)
	return p
}

func testParcelSendReceive(t *testing.T, protf func(io.ReadWriter) Protocol) {
	parcels := make([]*Parcel, 128)
	for i := range parcels {
		parcels[i] = testRandomParcel()
	}

	A, B := net.Pipe()
	defer A.Close()
	defer B.Close()
	sender := protf(A)
	receiver := protf(B)

	dl := time.Now().Add(time.Millisecond * 500)
	A.SetDeadline(dl)
	B.SetDeadline(dl)

	go func() {
		for _, p := range parcels {
			if err := sender.Send(p); err != nil {
				t.Errorf("prot %s: error sending %+v: %v", sender, p, err)
			}
		}
	}()

	for i := range parcels {
		p, err := receiver.Receive()
		if err != nil {
			t.Errorf("prot %s: error receiving %+v: %v", receiver, parcels[i], p)
			continue
		}

		if !reflect.DeepEqual(parcels[i], p) { // test parcels are made without an address
			t.Errorf("prot %s: received wrong message. want = %+v, got = %+v", receiver, parcels[i], p)
		}
	}
}

func testProtV9(rw io.ReadWriter) Protocol {
	decoder := gob.NewDecoder(rw)
	encoder := gob.NewEncoder(rw)
	return newProtocolV9(TestNet, 0x666, "9999", decoder, encoder)
}

func testProtV10(rw io.ReadWriter) Protocol {
	decoder := gob.NewDecoder(rw)
	encoder := gob.NewEncoder(rw)
	return newProtocolV10(decoder, encoder)
}

func testProtV11(rw io.ReadWriter) Protocol {
	return newProtocolV11(rw)
}

func TestParcelSendReceive(t *testing.T) {
	testParcelSendReceive(t, testProtV9)
	testParcelSendReceive(t, testProtV10)
	testParcelSendReceive(t, testProtV11)
}

func testHandshake(t *testing.T, protf func(io.ReadWriter) Protocol) {
	A, B := net.Pipe()
	defer A.Close()
	defer B.Close()

	sender := protf(A)
	receiver := protf(B)

	hs := testRandomHandshake()
	parcel := testRandomParcel()

	go func() {
		if err := sender.SendHandshake(hs); err != nil {
			t.Errorf("prot %s: send handshake err: %v", sender, err)
		}
		if err := sender.Send(parcel); err != nil {
			t.Errorf("prot %s: send parcel-after-handshake err: %v", sender, err)
		}
	}()

	if reply, err := receiver.ReadHandshake(); err != nil {
		t.Errorf("prot %s: read handshake err: %v", receiver, err)
	} else {

		if testEqualEndpointList(reply.Alternatives, hs.Alternatives) {
			// the order isn't guaranteed to be consistent, so check this first
			// then set to nil for reflect.DeepEqual to check the rest
			reply.Alternatives = nil
			hs.Alternatives = nil
		} else {
			t.Errorf("prot %s: alternatives different. sent = %v, got = %v", receiver, reply.Alternatives, hs.Alternatives)
		}

		if !reflect.DeepEqual(hs, reply) {
			t.Errorf("prot %s: handshake different. sent = %+v, got = %+v", receiver, hs, reply)
		}
	}

	if reply, err := receiver.Receive(); err != nil {
		t.Errorf("prot %s: receive parcel-after-handshake err: %v", receiver, err)
	} else if !reflect.DeepEqual(reply, parcel) {
		t.Errorf("prot %s: parcel-after-handshake not equal.", receiver)
	}
}

func TestProtocol_Handshake(t *testing.T) {
	for i := 0; i < 128; i++ {
		testHandshake(t, testProtV9)
		testHandshake(t, testProtV11)
	}
}

func TestProtocol_PeerShare(t *testing.T) {
	shares := make([][]Endpoint, 128)
	for i := range shares {
		shares[i] = testRandomEndpointList(rand.Intn(64))
	}

	for _, prot := range []Protocol{testProtV9(nil), testProtV10(nil), testProtV11(nil)} {
		for _, share := range shares {
			res, err := prot.MakePeerShare(share)
			if err != nil {
				t.Errorf("prot %d: %v", prot, err)
			}
			reverse, err := prot.ParsePeerShare(res)
			if err != nil {
				t.Errorf("prot %d: %v", prot, err)
			}

			if !testEqualEndpointList(reverse, share) {
				t.Errorf("prot %d: shares aren't the same. got = %v, want = %v", prot, reverse, share)
			}
		}
	}
}
