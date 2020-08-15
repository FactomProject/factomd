package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/util"
)

func createECEntriesfromBlocks(fBlock interfaces.IFBlock, eBlocks []*entryBlock.EBlock, height int) []interfaces.IECBlockEntry {
	ecEntries := []interfaces.IECBlockEntry{}
	ecEntries = append(ecEntries, entryCreditBlock.NewServerIndexNumber2(uint8(height%10+1)))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(1))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(2))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(3))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(4))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(5))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(6))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(7))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(8))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(9))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber(10))

	trans := fBlock.GetTransactions()
	for _, t := range trans {
		ecOut := t.GetECOutputs()
		for i, ec := range ecOut {
			increase := new(entryCreditBlock.IncreaseBalance)
			increase.ECPubKey = primitives.Byte32ToByteSlice32(ec.GetAddress().Fixed())
			increase.TXID = t.GetHash()
			increase.Index = uint64(i)
			increase.NumEC = ec.GetAmount() / fBlock.GetExchRate()
			ecEntries = append(ecEntries, increase)
		}
	}
	for _, eBlock := range eBlocks {
		if height == 0 {
			ecEntries = append(ecEntries, NewCommitChain(eBlock))
		} else {
			ecEntries = append(ecEntries, NewCommitEntry(eBlock))
		}
	}

	return ecEntries
}

func NewCommitEntry(eBlock *entryBlock.EBlock) *entryCreditBlock.CommitEntry {
	commit := entryCreditBlock.NewCommitEntry()

	commit.Version = 1
	err := commit.MilliTime.UnmarshalBinary([]byte{0, 0, 0, 0, 0, byte(eBlock.GetHeader().GetDBHeight())})
	if err != nil {
		panic(err)
	}
	commit.EntryHash = eBlock.Body.EBEntries[0]

	bin, err := commit.MarshalBinary()
	if err != nil {
		panic(err)
	}
	cost, err := util.EntryCost(bin)
	if err != nil {
		panic(err)
	}
	commit.Credits = cost

	SignCommit(0, commit)

	return commit
}

func NewCommitChain(eBlock *entryBlock.EBlock) *entryCreditBlock.CommitChain {
	commit := entryCreditBlock.NewCommitChain()

	commit.Version = 1
	err := commit.MilliTime.UnmarshalBinary([]byte{0, 0, 0, 0, 0, byte(eBlock.GetHeader().GetDBHeight())})
	if err != nil {
		panic(err)
	}
	commit.ChainIDHash = eBlock.GetHashOfChainIDHash()
	w := primitives.NewZeroHash()
	eh0 := eBlock.GetEntryHashes()[0].Bytes()
	cid := eBlock.GetHeader().GetChainID().Bytes()
	w.SetBytes(primitives.DoubleSha(append(eh0, cid...)))
	commit.Weld = w

	commit.EntryHash = eBlock.Body.EBEntries[0]
	bin, err := commit.MarshalBinary()
	if err != nil {
		panic(err)
	}
	cost, err := util.EntryCost(bin)
	if err != nil {
		panic(err)
	}
	commit.Credits = cost

	SignCommit(0, commit)

	return commit
}

func CreateTestEntryCreditBlock(prev interfaces.IEntryCreditBlock) interfaces.IEntryCreditBlock {
	block, err := entryCreditBlock.NextECBlock(prev)
	if err != nil {
		panic(err)
	}
	return block
}

func SignCommit(n uint64, tx interfaces.ISignable) {
	err := tx.Sign(NewPrivKey(n))
	if err != nil {
		panic(err)
	}
}
