package eventservices

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapFactoidBlock(t *testing.T) {
	block := factoid.NewFBlock(nil)
	block.AddTransaction(newTestTransaction())
	factoidBlock := mapFactoidBlock(block)

	assert.NotNil(t, factoidBlock)
}

func TestMapTransactions(t *testing.T) {

}

func TestMapTransaction(t *testing.T) {
	factoidTransaction := newTestTransaction()
	transaction := mapTransaction(factoidTransaction)

	assert.NotNil(t, transaction)
	assert.Equal(t, factoidTransaction.Txid.Bytes(), transaction.TransactionID)
	assert.Equal(t, factoidTransaction.BlockHeight, transaction.BlockHeight)
	assert.Equal(t, factoidTransaction.MilliTimestamp, transaction.Timestamp.Nanos)
}

func TestMapTransactionAddresses(t *testing.T) {
	address := factoid.RandomTransAddress()
	addresses := []interfaces.ITransAddress{address}
	transactionAddresses := mapTransactionAddresses(addresses)

	assert.NotNil(t, transactionAddresses)
	if assert.Equal(t, 1, len(transactionAddresses)) {
		assert.EqualValues(t, address.GetAddress().Bytes(), transactionAddresses[0].Address)
		assert.EqualValues(t, address.GetAmount(), transactionAddresses[0].Amount)
	}
}

func TestMapTransactionAddress(t *testing.T) {
	address := factoid.RandomTransAddress()
	transactionAddress := mapTransactionAddress(address)

	assert.NotNil(t, transactionAddress)
	assert.EqualValues(t, address.GetAddress().Bytes(), transactionAddress.Address)
	assert.EqualValues(t, address.GetAmount(), transactionAddress.Amount)
}

func TestMapRCDs(t *testing.T) {

}

func TestMapRCD(t *testing.T) {
	factoidRCD := &factoid.RCD_1{
		PublicKey: [32]byte{},
	}

	rcd := mapRCD(factoidRCD)

	assert.NotNil(t, rcd)
	assert.NotNil(t, rcd.Rcd)
	assert.EqualValues(t, make([]byte, 32), rcd.Rcd)
}

func TestMapSignatureBlocks(t *testing.T) {
	priv := make([]byte, constants.SIGNATURE_LENGTH)
	signature := make([]byte, constants.SIGNATURE_LENGTH)
	block := factoid.NewSingleSignatureBlock(priv, signature)
	blocks := []interfaces.ISignatureBlock{block}

	signatureBlocks := mapSignatureBlocks(blocks)

	assert.NotNil(t, signatureBlocks)
	if assert.Equal(t, 1, len(signatureBlocks)) {
		if assert.Equal(t, 1, len(signatureBlocks[0].Signature)) {
			assert.EqualValues(t, constants.SIGNATURE_LENGTH, len(signatureBlocks[0].Signature[0]))
		}
	}
}

func TestMapSignatureBlock(t *testing.T) {
	priv := make([]byte, constants.SIGNATURE_LENGTH)
	signature := make([]byte, constants.SIGNATURE_LENGTH)
	block := factoid.NewSingleSignatureBlock(priv, signature)

	signatureBlock := mapSignatureBlock(block)

	assert.NotNil(t, signatureBlock)
	assert.NotNil(t, signatureBlock.Signature)
	if assert.Equal(t, 1, len(signatureBlock.Signature)) {
		assert.EqualValues(t, constants.SIGNATURE_LENGTH, len(signatureBlock.Signature[0]))
	}
}

func TestMapFactoidSignatureBlockSingatures(t *testing.T) {
	signature := new(factoid.FactoidSignature)
	signature.UnmarshalBinary([]byte(""))
	signatures := []interfaces.ISignature{signature}

	factoidSignatures := mapFactoidSignatureBlockSignatures(signatures)

	assert.NotNil(t, factoidSignatures)
	if assert.Equal(t, 1, len(factoidSignatures)) {
		expected := make([]uint8, constants.SIGNATURE_LENGTH)
		assert.EqualValues(t, expected, factoidSignatures[0])
	}
}

func newTestTransaction() *factoid.Transaction {
	tx := new(factoid.Transaction)
	tx.AddOutput(factoid.NewAddress([]byte("")), 1)
	tx.SetTimestamp(primitives.NewTimestampFromSeconds(60 * 10 * uint32(1)))
	return tx
}
