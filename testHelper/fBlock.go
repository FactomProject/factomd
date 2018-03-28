package testHelper

//A package for functions used multiple times in tests that aren't useful in production code.

import (
	"encoding/hex"

	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func CreateTestFactoidBlock(prev interfaces.IFBlock) interfaces.IFBlock {
	fBlock := CreateTestFactoidBlockWithCoinbase(prev, NewFactoidAddress(0), DefaultCoinbaseAmount)

	ecTx := new(factoid.Transaction)
	ecTx.AddInput(NewFactoidAddress(0), fBlock.GetExchRate()*100)
	ecTx.AddECOutput(NewECAddress(0), fBlock.GetExchRate()*100)
	ecTx.SetTimestamp(primitives.NewTimestampFromSeconds(60 * 10 * uint32(fBlock.GetDBHeight())))

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

func CreateTestFactoidBlockWithTransaction(prev interfaces.IFBlock, sentSecret string, receivePublic []byte, amt uint64) interfaces.IFBlock {
	fBlock := CreateTestFactoidBlockWithCoinbase(prev, NewFactoidAddress(0), DefaultCoinbaseAmount)

	for i := 0; i < 5; i++ {
		err := fBlock.AddTransaction(newTrans(fBlock.GetDatabaseHeight(), sentSecret, receivePublic, amt))
		if err != nil {
			panic(err)
		}
	}

	return fBlock
}

func newTrans(height uint32, sentSecret string, receivePublic []byte, amt uint64) *factoid.Transaction {
	privKey, pubKey, _, err := factoid.PrivateKeyStringToEverythingString(sentSecret)
	if err != nil {
		panic(err)
	}

	privBytes, err := hex.DecodeString(privKey)
	if err != nil {
		panic(err)
	}

	/*	pubBytes, err := hex.DecodeString(pubKey)
		if err != nil {
			panic(err)
		}
	*/
	add, err := factoid.PublicKeyStringToFactoidAddress(pubKey)
	if err != nil {
		panic(err)
	}

	ecTx := new(factoid.Transaction)
	ecTx.AddInput(add, amt)
	ecTx.AddOutput(factoid.NewAddress(receivePublic), amt)
	ecTx.SetTimestamp(primitives.NewTimestampFromSeconds(60 * 10 * uint32(height)))

	fee, err := ecTx.CalculateFee(1000)
	if err != nil {
		panic(err)
	}
	in, err := ecTx.GetInput(0)
	if err != nil {
		panic(err)
	}
	in.SetAmount(in.GetAmount() + fee*2)

	// SIGN
	rcd, err := factoid.PublicKeyStringToFactoidRCDAddress(pubKey)
	if err != nil {
		panic(err)
	}

	ecTx.AddAuthorization(rcd)
	data, err := ecTx.MarshalBinarySig()
	if err != nil {
		panic(err)
	}

	sig := factoid.NewSingleSignatureBlock(privBytes, data)

	//str, err := sig.JSONString()

	//fmt.Printf("sig, err - %v, %v\n", str, err)

	ecTx.SetSignatureBlock(0, sig)

	err = ecTx.Validate(1)
	if err != nil {
		panic(err)
	}

	err = ecTx.ValidateSignatures()
	if err != nil {
		panic(err)
	}

	// END SIGN

	return ecTx
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
	block := factoid.NewFBlock(prev)
	tx := new(factoid.Transaction)
	tx.AddOutput(address, amount)
	tx.SetTimestamp(primitives.NewTimestampFromSeconds(60 * 10 * uint32(block.GetDBHeight())))
	err := block.AddCoinbase(tx)
	if err != nil {
		panic(err)
	}
	return block
}
