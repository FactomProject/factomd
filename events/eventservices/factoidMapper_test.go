package eventservices

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/stretchr/testify/assert"
)

func TestMapFactoidBlock(t *testing.T) {
	block := factoid.NewFBlock(nil)
	factoidBlock := mapFactoidBlock(block)

	assert.NotNil(t, factoidBlock)
	assert.NotNil(t, factoidBlock.BlockHeight)
	assert.NotNil(t, factoidBlock.BodyMerkleRoot)
	assert.NotNil(t, factoidBlock.KeyMerkleRoot)
	assert.NotNil(t, factoidBlock.PreviousKeyMerkleRoot)
	assert.NotNil(t, factoidBlock.ExchangeRate)
	assert.NotNil(t, factoidBlock.BlockHeight)
	assert.GreaterOrEqual(t, uint32(1), factoidBlock.TransactionCount)
	assert.NotNil(t, factoidBlock.Transactions)
}

func TestMapTransaction(t *testing.T) {
	factoidTransaction := newTestTransaction()
	transaction := mapTransaction(factoidTransaction, 1)

	assert.NotNil(t, transaction)
	assert.Equal(t, factoidTransaction.GetSigHash().Bytes(), transaction.TransactionID)
	assert.Equal(t, factoidTransaction.BlockHeight, transaction.BlockHeight)
	assert.Equal(t, uint32(1), transaction.MinuteNumber)
	assert.Equal(t, int64(factoidTransaction.MilliTimestamp/1000), transaction.Timestamp.Seconds)
	assert.Equal(t, int32(factoidTransaction.MilliTimestamp%1000), transaction.Timestamp.Nanos)
}

func TestMapTransactions(t *testing.T) {
	factoidTransaction := newTestTransaction()
	factoidTransactions := []interfaces.ITransaction{factoidTransaction}
	var endOfPeriod [10]int
	for i := 0; i < len(endOfPeriod); i++ {
		endOfPeriod[i] = len(factoidTransactions)
	}
	transactions := mapTransactions(factoidTransactions, endOfPeriod)

	assert.NotNil(t, transactions)
	if assert.Equal(t, 1, len(transactions)) {
		assert.Equal(t, factoidTransaction.GetSigHash().Bytes(), transactions[0].TransactionID)
		assert.Equal(t, factoidTransaction.BlockHeight, transactions[0].BlockHeight)
		assert.Equal(t, int64(factoidTransaction.MilliTimestamp/1000), transactions[0].Timestamp.Seconds)
		assert.Equal(t, int32(factoidTransaction.MilliTimestamp%1000), transactions[0].Timestamp.Nanos)
	}
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
	publicKey := [constants.ADDRESS_LENGTH]byte{}
	rcd := factoid.NewRCD_1(publicKey[:])
	rcds := []interfaces.IRCD{rcd}
	mappedRCDs := mapRCDs(rcds)

	assert.NotNil(t, mappedRCDs)
	if assert.Equal(t, 1, len(mappedRCDs)) {
		if assert.NotNil(t, mappedRCDs[0].Rcd) {
			assert.Equal(t, constants.ADDRESS_LENGTH, len(mappedRCDs[0].GetRcd1().PublicKey))
			assert.EqualValues(t, publicKey[:], mappedRCDs[0].GetRcd1().PublicKey)
		}
	}
}

func TestMapRCD(t *testing.T) {
	publicKey := [constants.ADDRESS_LENGTH]byte{}
	rcd := factoid.NewRCD_1(publicKey[:])
	mappedRCD := mapRCD(rcd)

	assert.NotNil(t, mappedRCD)
	if assert.NotNil(t, mappedRCD.Rcd) {
		assert.Equal(t, constants.ADDRESS_LENGTH, len(mappedRCD.GetRcd1().PublicKey))
		assert.EqualValues(t, publicKey[:], mappedRCD.GetRcd1().PublicKey)
	}
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
	address := factoid.NewAddress([]byte(""))
	tx := new(factoid.Transaction)
	tx.AddInput(address, 2)
	tx.AddOutput(address, 1)
	tx.AddOutput(address, 1)
	tx.SetTimestamp(primitives.NewTimestampFromSeconds(60 * 10 * uint32(1)))
	return tx
}
