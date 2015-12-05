// Copyright (c) 2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"errors"

	"github.com/FactomProject/factomd/btcd/wire"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Checkpoint identifies a known good point in the block chain.  Using
// checkpoints allows a few optimizations for old blocks during initial download
// and also prevents forks from old blocks.
//
// Each checkpoint is selected based upon several factors.  See the
// documentation for blockchain.IsCheckpointCandidate for details on the
// selection criteria.
type Checkpoint struct {
	Height int64
	Hash   interfaces.IHash
}

// Params defines a Bitcoin network by its parameters.  These parameters may be
// used by Bitcoin applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
type Params struct {
	Name        string
	Net         wire.BitcoinNet
	DefaultPort string
}

// MainNetParams defines the network parameters for the main Bitcoin network.
var MainNetParams = Params{
	Name:        "mainnet",
	Net:         wire.MainNet,
	DefaultPort: "12204", //"8108",
}

// RegressionNetParams defines the network parameters for the regression test
// Bitcoin network.  Not to be confused with the test Bitcoin network (version
// 3), this network is sometimes simply called "testnet".
var RegressionNetParams = Params{
	Name: "devnet",
	Net:  wire.TestNet,
	DefaultPort: "12204",		//"18444",
}

// TestNet3Params defines the network parameters for the test Bitcoin network
// (version 3).  Not to be confused with the regression test network, this
// network is sometimes simply called "testnet".
var TestNet3Params = Params{
	Name:        "testnet3",
	Net:         wire.TestNet3,
	DefaultPort: "18108",
}

var (
	// ErrDuplicateNet describes an error where the parameters for a Bitcoin
	// network could not be set due to the network already being a standard
	// network or previously-registered into this package.
	ErrDuplicateNet = errors.New("duplicate Bitcoin network")
)

var (
	registeredNets = map[wire.BitcoinNet]struct{}{
		MainNetParams.Net:       struct{}{},
		TestNet3Params.Net:      struct{}{},
		RegressionNetParams.Net: struct{}{},
	}
)

// Register registers the network parameters for a Bitcoin network.  This may
// error with ErrDuplicateNet if the network is already registered (either
// due to a previous Register call, or the network being one of the default
// networks).
//
// Network parameters should be registered into this package by a main package
// as early as possible.  Then, library packages may lookup networks or network
// parameters based on inputs and work regardless of the network being standard
// or not.
func Register(params *Params) error {
	if _, ok := registeredNets[params.Net]; ok {
		return ErrDuplicateNet
	}
	registeredNets[params.Net] = struct{}{}
	return nil
}

// newinterfaces.IHashFromStr converts the passed big-endian hex string into a
// interfaces.IHash.  It only differs from the one available in wire in that
// it panics on an error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newIHashFromStr(hexStr string) interfaces.IHash {
	sha, err := primitives.NewShaHashFromStr(hexStr)
	if err != nil {
		// Ordinarily I don't like panics in library code since it
		// can take applications down without them having a chance to
		// recover which is extremely annoying, however an exception is
		// being made in this case because the only way this can panic
		// is if there is an error in the hard-coded hashes.  Thus it
		// will only ever potentially panic on init and therefore is
		// 100% predictable.
		panic(err)
	}
	return sha
}
