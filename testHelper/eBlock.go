package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"encoding/hex"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

func CreateTestEntryBlock(prev *entryBlock.EBlock) *entryBlock.EBlock {
	e := entryBlock.NewEBlock()
	entryStr := "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a1974625f25d9375533b44505964af993212ef7c13314736b2c76a37c73571d89d8b21c6180f7430677d46d93a3e17b68e6a25dc89ecc092cee1459101578859f7f6969d171a092a1d04f067d55628b461c6a106b76b4bc860445f87b0052cdc5f2bfd000002d800001b080000000272d72e71fdee4984ecb30eedcc89cb171d1f5f02bf9a8f10a8b2cfbaf03efe1c0000000000000000000000000000000000000000000000000000000000000001"
	h, err := hex.DecodeString(entryStr)
	if err != nil {
		panic(err)
	}
	err = e.UnmarshalBinary(h)
	if err != nil {
		panic(err)
	}

	if prev != nil {
		keyMR, err := prev.KeyMR()
		if err != nil {
			panic(err)
		}

		e.Header.SetPrevKeyMR(keyMR)
		hash, err := prev.Hash()
		if err != nil {
			panic(err)
		}
		e.Header.SetPrevLedgerKeyMR(hash)
		e.Header.SetDBHeight(prev.Header.GetDBHeight() + 1)

		e.Header.SetChainID(prev.Header.GetChainID())
	} else {
		e.Header.SetPrevKeyMR(primitives.NewZeroHash())
		e.Header.SetDBHeight(0)
	}

	return e
}
