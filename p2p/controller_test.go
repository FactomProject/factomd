package p2p_test

import (
	"math/rand"
	"testing"

	"github.com/PaulSnow/factom2d/common/primitives/random"
	"github.com/PaulSnow/factom2d/p2p"
)

func BenchmarkBroadcastWithSliceBuild(b *testing.B) {
	conns := make(map[string]*p2p.Peer)
	for i := 0; i < 100; i++ {
		con := new(p2p.Peer)
		con.Hash = random.RandomString()
		conns[con.Hash] = con
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// Make list
		regularPeers := make([]string, 0)

		for peerHash, _ := range conns {
			regularPeers = append(regularPeers, peerHash)
		}

		rand.Shuffle(len(regularPeers), func(i, j int) {
			regularPeers[i], regularPeers[j] = regularPeers[j], regularPeers[i]
		})

		for peerHash := range regularPeers {
			var _ = peerHash
		}
	}
}

func BenchmarkBroadcastWithoutSliceBuild(b *testing.B) {
	conns := make(map[string]*p2p.Peer)
	for i := 0; i < 100; i++ {
		con := new(p2p.Peer)
		con.Hash = random.RandomString()
		conns[con.Hash] = con
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// Make list

		for peerHash, _ := range conns {
			var _ = peerHash
		}

	}
}
