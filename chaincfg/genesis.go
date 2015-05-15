// Copyright (c) 2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"github.com/FactomProject/btcd/wire"
)

// buyer #1
var rcdbuy1 = wire.RCDHash([wire.RCDHashSize]byte{ // Make go vet happy.
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x28, 0xc3, 0x4f, 0x3a, 0x5e, 0x33, 0x2a, 0x1f, 0xc7, 0xb2, 0xb7, 0x3c, 0xf1, 0x88, 0x91, 0x0f,
})

// buyer #2
var rcdbuy2 = wire.RCDHash([wire.RCDHashSize]byte{ // Make go vet happy.
	0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
	0x28, 0xc3, 0x4f, 0x3a, 0x5e, 0x33, 0x2a, 0x1f, 0xc7, 0xb2, 0xb7, 0x3c, 0xf1, 0x88, 0x91, 0x0f,
})

// buyer #3
var rcdbuy3 = wire.RCDHash([wire.RCDHashSize]byte{ // Make go vet happy.
	0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03,
	0x28, 0xc3, 0x4f, 0x3a, 0x5e, 0x33, 0x2a, 0x1f, 0xc7, 0xb2, 0xb7, 0x3c, 0xf1, 0x88, 0x91, 0x0f,
})

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network (version 3).
var genesisCoinbaseTx = wire.MsgTx{
	// Version: 1,
	//	Version: 5, // FIXME
	Version: 0,
	//	LockTime: 0x12345000000, // FIXME, testing
	LockTime: 0,
	TxOut: []*wire.TxOut{
		{
			//			Value: 0x12a05f200, // coinbase
			Value: 50 * 1e8,
		},

		{
			Value:   100 * 1e8,
			RCDHash: rcdbuy1,
		},
		{
			Value:   200 * 1e8,
			RCDHash: rcdbuy2,
		},
		{
			Value:   300 * 1e8,
			RCDHash: rcdbuy3,
		},
	},
	TxIn: []*wire.TxIn{
		{
			PreviousOutPoint: wire.OutPoint{
				Hash:  wire.ShaHash{},
				Index: 0xffffffff,
			},
		},
	},
}

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = wire.ShaHash([wire.HashSize]byte{ // Make go vet happy.
	/*
		0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
		0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
		0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
		0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
	*/

	/*
		0x04,
		0x9c,
		0xc1,
		0xa5,
		0xb8,
		0xcf,
		0xe0,
		0x1b,
		0x88,
		0x65,
		0x9f,
		0x9e,
		0xe7,
		0x64,
		0x1c,
		0xaa,
		0x69,
		0x18,
		0x25,
		0x33,
		0xde,
		0x0b,
		0xb8,
		0xc9,
		0xe6,
		0x10,
		0x3c,
		0xc5,
		0x09,
		0x1d,
		0x86,
		0xc4,
	*/

	0x4b,
	0xe7,
	0x57,
	0x0e,
	0x8f,
	0x70,
	0xeb,
	0x09,
	0x36,
	0x40,
	0xc8,
	0x46,
	0x82,
	0x74,
	0xba,
	0x75,
	0x97,
	0x45,
	0xa7,
	0xaa,
	0x2b,
	0x7d,
	0x25,
	0xab,
	0x1e,
	0x04,
	0x21,
	0xb2,
	0x59,
	0x84,
	0x50,
	0x14,
})

/*
var genesisHash = wire.ShaHash([wire.HashSize]byte{ // Make go vet happy.
	0xc6, 0xbe, 0x35, 0x47, 0x9f, 0xe9, 0xe0, 0x0a, 0x7e, 0xae, 0x97, 0xd2, 0x2c, 0x20, 0xea, 0x28,
	0x7a, 0x21, 0xb5, 0x13, 0x37, 0x7b, 0x39, 0x97, 0xc1, 0x53, 0x0d, 0xe5, 0x99, 0x8f, 0xb8, 0xe4,
})
*/

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = wire.ShaHash([wire.HashSize]byte{ // Make go vet happy.
	/*
		0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2,
		0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
		0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
		0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a,
	*/
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.FBlockHeader{

		PrevBlock:  wire.ShaHash{},    // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: genesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		PrevHash3:  wire.Sha3Hash{},
        ExchRate:   uint64(666666),
        DBHeight:   uint32(0),
        TransCnt:   uint64(1),
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = wire.ShaHash([wire.HashSize]byte{ // Make go vet happy.
	0x06, 0x22, 0x6e, 0x46, 0x11, 0x1a, 0x0b, 0x59,
	0xca, 0xaf, 0x12, 0x60, 0x43, 0xeb, 0x5b, 0xbf,
	0x28, 0xc3, 0x4f, 0x3a, 0x5e, 0x33, 0x2a, 0x1f,
	0xc7, 0xb2, 0xb7, 0x3c, 0xf1, 0x88, 0x91, 0x0f,
})

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the regression test network.  It is the same as the merkle root for
// the main network.
var regTestGenesisMerkleRoot = genesisMerkleRoot

// regTestGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the regression test network.
var regTestGenesisBlock = wire.MsgBlock{
	Header: wire.FBlockHeader{

		PrevBlock:  wire.ShaHash{},           // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: regTestGenesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		PrevHash3:  wire.Sha3Hash{},
        ExchRate:   uint64(666666),
        DBHeight:   uint32(0),
        TransCnt:   uint64(1),
    },
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = wire.ShaHash([wire.HashSize]byte{ // Make go vet happy.
	0x43, 0x49, 0x7f, 0xd7, 0xf8, 0x26, 0x95, 0x71,
	0x08, 0xf4, 0xa3, 0x0f, 0xd9, 0xce, 0xc3, 0xae,
	0xba, 0x79, 0x97, 0x20, 0x84, 0xe9, 0x0e, 0xad,
	0x01, 0xea, 0x33, 0x09, 0x00, 0x00, 0x00, 0x00,
})

// testNet3GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 3).  It is the same as the merkle root
// for the main network.
var testNet3GenesisMerkleRoot = genesisMerkleRoot

// testNet3GenesisBlock defines the genesis block of the block chain which
// serves as the public transaction ledger for the test network (version 3).
var testNet3GenesisBlock = wire.MsgBlock{
	Header: wire.FBlockHeader{
		PrevBlock:  wire.ShaHash{},            // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: testNet3GenesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		PrevHash3:  wire.Sha3Hash{},
        ExchRate:   uint64(666666),
        DBHeight:   uint32(0),
        TransCnt:   uint64(1),
    },
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = wire.ShaHash([wire.HashSize]byte{ // Make go vet happy.
	0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a,
	0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
	0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
	0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
})

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the simulation test network.  It is the same as the merkle root for
// the main network.
var simNetGenesisMerkleRoot = genesisMerkleRoot

// simNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the simulation test network.
var simNetGenesisBlock = wire.MsgBlock{
	Header: wire.FBlockHeader{
		PrevBlock:  wire.ShaHash{},          // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: simNetGenesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		PrevHash3:  wire.Sha3Hash{},
        ExchRate:   uint64(666666),
        DBHeight:   uint32(0),
        TransCnt:   uint64(1),
    },
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}
