package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func createECEntriesfromFBlock(fBlock interfaces.IFBlock, height int) []interfaces.IECBlockEntry {
	ecEntries := []interfaces.IECBlockEntry{}
	ecEntries = append(ecEntries, entryCreditBlock.NewServerIndexNumber2(uint8(height%10+1)))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(0))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(1))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(2))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(3))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(4))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(5))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(6))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(7))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(8))
	ecEntries = append(ecEntries, entryCreditBlock.NewMinuteNumber2(9))

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

	if height == 0 {

	} else {

	}

	return ecEntries
}

func NewCommitChain(eBlock *entryBlock.EBlock) *entryCreditBlock.CommitChain {
	commit := entryCreditBlock.NewCommitChain()

	commit.Version = 1
	err := commit.MilliTime.UnmarshalBinary([]byte{0, 0, 0, 0, 0, byte(eBlock.GetHeader().GetDBHeight())})
	if err != nil {
		panic(err)
	}
	commit.ChainIDHash = eBlock.GetHashOfChainIDHash()
	commit.Weld = eBlock.GetWeldHash()
	commit.EntryHash = eBlock.Body.EBEntries[0]
	/*
		commit.Credits = 0
		commit.ECPubKey = nil
		commit.Sig = nil
	*/
	return commit
}

func CreateTestEntryCreditBlock(prev interfaces.IEntryCreditBlock) interfaces.IEntryCreditBlock {
	block, err := entryCreditBlock.NextECBlock(prev)
	if err != nil {
		panic(err)
	}
	return block
}
