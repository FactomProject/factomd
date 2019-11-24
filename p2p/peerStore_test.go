package p2p

import (
	"fmt"
	"reflect"
	"testing"
)

func testStore() *PeerStore {
	ps := NewPeerStore()
	return ps
}

func testPeer(addr, port string, id uint32, incoming bool) *Peer {
	p := new(Peer)
	p.Endpoint.IP = addr
	p.Endpoint.Port = port
	p.IsIncoming = incoming
	p.Hash = fmt.Sprintf("%s:%s %08x", addr, port, id)
	return p
}

const (
	incoming = 1
	outgoing = 4
	total    = 5
	unique   = 3 // unique ips
)

func testPeers() []*Peer {
	return []*Peer{
		testPeer("127.0.0.1", "8088", 1, false),
		testPeer("127.0.0.1", "8088", 2, false), // different hash by id
		testPeer("127.0.0.1", "8089", 1, false), // different hash by port
		testPeer("127.0.0.2", "8088", 1, false), // different hash by address
		testPeer("127.0.0.3", "8088", 3, true),  // incoming
	}
}

func testAll() (*PeerStore, []*Peer) {
	ps := testStore()
	peers := testPeers()
	for _, p := range peers {
		ps.Add(p)
	}
	return ps, peers
}

func TestPeerStore_Add(t *testing.T) {
	ps := testStore()
	peers := testPeers()

	type args struct {
		p *Peer
	}
	tests := []struct {
		name    string
		ps      *PeerStore
		args    args
		wantErr bool
	}{
		{"normal", ps, args{peers[0]}, false},
		{"duplicate", ps, args{peers[0]}, true},
		{"diff id", ps, args{peers[1]}, false},
		{"diff hash", ps, args{peers[2]}, false},
		{"diff add", ps, args{peers[3]}, false},
		{"nil pointer", ps, args{nil}, true},
		{"incoming", ps, args{peers[4]}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ps.Add(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("PeerStore.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	found := 0
	for _, p := range ps.peers {
		for _, p2 := range peers {
			if p == p2 {
				found++
				break
			}
		}
	}
	if found != len(peers) {
		t.Errorf("not all peers found, missing %d", len(peers)-found)
	}
}

func TestPeerStore_Remove(t *testing.T) {
	ps := testStore()
	p1 := testPeer("127.0.0.1", "8088", 1, false)
	p2 := testPeer("127.0.0.1", "8088", 1, false)
	p3 := testPeer("127.0.0.2", "8088", 1, false)
	ps.Add(p1)
	ps.Add(p3)
	ps.Remove(p2)
	ps.Remove(nil)
	if _, ok := ps.peers[p1.Hash]; !ok {
		t.Error("Removed peer p1 despite not being the same pointer")
	}
	ps.Remove(p1)
	if len(ps.peers) != 1 {
		t.Error("inconsistent peer count")
	}
	ps.Remove(p3)
	if len(ps.peers) != 0 {
		t.Error("inconsistent peer count 2")
	}
}

func TestPeerStore_Total(t *testing.T) {
	ps, _ := testAll()
	if ps.Total() != len(ps.peers) {
		t.Error("total reported the wrong count compared to map")
	}
	if ps.Total() != total {
		t.Error("total reported the wrong count compared to how many we added")
	}
}

func TestPeerStore_Outgoing(t *testing.T) {
	ps, _ := testAll()

	if ps.Outgoing() != outgoing {
		t.Errorf("Outgoing reported the wrong count. is %d should be %d", ps.Outgoing(), outgoing)
	}
}

func TestPeerStore_Incoming(t *testing.T) {
	ps, _ := testAll()

	if ps.Incoming() != incoming {
		t.Errorf("Incoming reported the wrong count. is %d should be %d", ps.Incoming(), incoming)
	}
}

func TestPeerStore_Get(t *testing.T) {
	ps, peers := testAll()

	type args struct {
		hash string
	}
	tests := []struct {
		name string
		ps   *PeerStore
		args args
		want *Peer
	}{
		{"get0", ps, args{peers[0].Hash}, peers[0]},
		{"get1", ps, args{peers[1].Hash}, peers[1]},
		{"get2", ps, args{peers[2].Hash}, peers[2]},
		{"get3", ps, args{peers[3].Hash}, peers[3]},
		{"get4", ps, args{peers[4].Hash}, peers[4]},
		{"nonexist", ps, args{"foo"}, nil},
		{"empty input", ps, args{""}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.Get(tt.args.hash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerStore_IsConnected(t *testing.T) {
	ps, _ := testAll()
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		ps   *PeerStore
		args args
		want bool
	}{
		{"test localhost", ps, args{"127.0.0.1"}, true},
		{"test localhost2", ps, args{"127.0.0.2"}, true},
		{"test nonexist", ps, args{"1.2.3.4"}, false},
		{"test empty input", ps, args{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.Connections(tt.args.addr); (got > 0) != tt.want {
				t.Errorf("PeerStore.IsConnected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerStore_Count(t *testing.T) {
	ps, _ := testAll()
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		ps   *PeerStore
		args args
		want int
	}{
		{"localhost.1", ps, args{"127.0.0.1"}, 3},
		{"localhost.2", ps, args{"127.0.0.2"}, 1},
		{"localhost.3", ps, args{"127.0.0.3"}, 1},
		{"nonexistent", ps, args{"foo"}, 0},
		{"empty input", ps, args{""}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.Count(tt.args.addr); got != tt.want {
				t.Errorf("PeerStore.Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func find(p *Peer, peers []*Peer) bool {
	if p == nil {
		return false
	}
	for _, s := range peers {
		if p == s {
			return true
		}
	}
	return false
}

func TestPeerStore_Slice(t *testing.T) {
	ps, peers := testAll()

	s1 := ps.Slice()
	s2 := ps.Slice()
	ps.Remove(peers[0])
	s3 := ps.Slice()

	for i := 0; i < total; i++ {
		if !find(s1[i], peers) {
			t.Errorf("data from s1 is inconsistent with list")
		}
		if !find(s2[i], peers) {
			t.Errorf("data from s2 is inconsistent with list")
		}
		if i < total-1 && !find(s3[i], peers) {
			t.Errorf("data from s3 is inconsistent with list")
		}

	}
}
