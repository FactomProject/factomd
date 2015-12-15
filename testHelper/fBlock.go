package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"github.com/FactomProject/factomd/common/factoid"
	factoidBlock "github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
)

func CreateTestFactoidBlock(prev interfaces.IFBlock) interfaces.IFBlock {
	fBlock := CreateTestFactoidBlockWithCoinbase(prev, NewFactoidAddress(0), DefaultCoinbaseAmount)

	ecTx := new(factoid.Transaction)
	ecTx.AddInput(NewFactoidAddress(0), fBlock.GetExchRate()*100)
	ecTx.AddECOutput(NewECAddress(0), fBlock.GetExchRate()*100)

	fee, err := ecTx.CalculateFee(1000)
	if err != nil {
		panic(err)
	}
	in, err := ecTx.GetInput(0)
	if err != nil {
		panic(err)
	}
	in.SetAmount(in.GetAmount() + fee)

	SignFactoidTransaction(0, ecTx)

	err = fBlock.AddTransaction(ecTx)
	if err != nil {
		panic(err)
	}

	return fBlock
}

func SignFactoidTransaction(n uint64, tx interfaces.ITransaction) {
	tx.AddAuthorization(NewFactoidRCDAddress(n))
	data, err := tx.MarshalBinarySig()
	if err != nil {
		panic(err)
	}

	sig := factoid.NewSingleSignatureBlock(NewPrivKey(n), data)

	//str, err := sig.JSONString()

	//fmt.Printf("sig, err - %v, %v\n", str, err)

	tx.SetSignatureBlock(0, sig)

	err = tx.Validate(1)
	if err != nil {
		panic(err)
	}

	err = tx.ValidateSignatures()
	if err != nil {
		panic(err)
	}
}

func CreateTestFactoidBlockWithCoinbase(prev interfaces.IFBlock, address interfaces.IAddress, amount uint64) interfaces.IFBlock {
	block := factoidBlock.NewFBlockFromPreviousBlock(1, prev)
	tx := new(factoid.Transaction)
	tx.AddOutput(address, amount)
	err := block.AddCoinbase(tx)
	if err != nil {
		panic(err)
	}
	return block
}
