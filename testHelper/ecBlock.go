package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
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

func NewCommitChain() *entryCreditBlock.CommitChain {
	commit := entryCreditBlock.NewCommitChain()

	/*
		if p, err := hex.DecodeString(c.CommitChainMsg); err != nil {
			wsLog.Error(err)
			ctx.WriteHeader(httpBad)
			ctx.Write([]byte(err.Error()))
			return
		} else {
			_, err := commit.UnmarshalBinaryData(p)
			if err != nil {
				wsLog.Error(err)
				ctx.WriteHeader(httpBad)
				ctx.Write([]byte(err.Error()))
				return
			}
		}

		if err := factomapi.CommitChain(commit); err != nil {
			wsLog.Error(err)
			ctx.WriteHeader(httpBad)
			ctx.Write([]byte(err.Error()))
			return
		}*/

	return commit
}

func CreateTestEntryCreditBlock(prev interfaces.IEntryCreditBlock) interfaces.IEntryCreditBlock {
	block, err := entryCreditBlock.NextECBlock(prev)
	if err != nil {
		panic(err)
	}
	return block
}
