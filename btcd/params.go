// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"github.com/FactomProject/factomd/common/messages"
)

// activeNetParams is a pointer to the parameters specific to the
// currently active bitcoin network.
var activeNetParams = &mainNetParams

// Params defines a Bitcoin network by its parameters.  These parameters may be
// used by Bitcoin applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
type Params struct {
	Name        string
	Net         messages.FactomNet
	DefaultPort string
}

// MainNetParams defines the network parameters for the main Bitcoin network.
var MainNetParams = Params{
	Name:        "mainnet",
	Net:         messages.MainNet,
	DefaultPort: "12204", //"8108",
}

// RegressionNetParams defines the network parameters for the regression test
// Bitcoin network.  Not to be confused with the test Bitcoin network (version
// 3), this network is sometimes simply called "testnet".
var RegressionNetParams = Params{
	Name: "devnet",
	Net:  messages.TestNet,
	DefaultPort: "12204",		//"18444",
}

// TestNet3Params defines the network parameters for the test Bitcoin network
// (version 3).  Not to be confused with the regression test network, this
// network is sometimes simply called "testnet".
var TestNet3Params = Params{
	Name:        "testnet3",
	Net:         messages.TestNet3,
	DefaultPort: "18108",
}

// params is used to group parameters for various networks such as the main
// network and test networks.
type params struct {
	*Params
	rpcPort  string
	dnsSeeds []string
}

// mainNetParams contains parameters specific to the main network
// (messages.MainNet).  NOTE: The RPC port is intentionally different than the
// reference implementation because btcd does not handle wallet requests.  The
// separate wallet process listens on the well-known port and forwards requests
// it does not handle on to btcd.  This approach allows the wallet process
// to emulate the full reference implementation RPC API.
var mainNetParams = params{
	Params:  &MainNetParams,
	rpcPort: "8384",
	dnsSeeds: []string{
		//		"factom.network",
		/*
			"seed.bitcoin.sipa.be",
			"dnsseed.bluematt.me",
			"dnsseed.bitcoin.dashjr.org",
			"seed.bitcoinstats.com",
			"seed.bitnodes.io",
			"bitseed.xf2.org",
			"seeds.bitcoin.open-nodes.org",
		*/
		"52.27.143.38",
	},
}

// regressionNetParams contains parameters specific to the regression test
// network (messages.TestNet).  NOTE: The RPC port is intentionally different
// than the reference implementation - see the mainNetParams comment for
// details.
var regressionNetParams = params{
	Params:   &RegressionNetParams,
	rpcPort:  "18334",
	dnsSeeds: []string{},
}

// netName returns the name used when referring to a bitcoin network.  At the
// time of writing, btcd currently places blocks for testnet version 3 in the
// data and log directory "testnet", which does not match the Name field of the
// chaincfg parameters.  This function can be used to override this directory
// name as "testnet" when the passed active network matches messages.TestNet3.
//
// A proper upgrade to move the data and log directories for this network to
// "testnet3" is planned for the future, at which point this function can be
// removed and the network parameter's name used instead.
func netName(chainParams *params) string {
	switch chainParams.Net {
	case messages.TestNet3:
		return "testnet"
	default:
		return chainParams.Name
	}
}
