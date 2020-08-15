package eventservices

import (
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/events/eventmessages/generated/eventmessages"
)

func mapFactoidBlock(block interfaces.IFBlock) *eventmessages.FactoidBlock {
	result := &eventmessages.FactoidBlock{
		BodyMerkleRoot:              block.GetBodyMR().Bytes(),
		KeyMerkleRoot:               block.GetKeyMR().Bytes(),
		PreviousKeyMerkleRoot:       block.GetPrevKeyMR().Bytes(),
		PreviousLedgerKeyMerkleRoot: block.GetLedgerKeyMR().Bytes(),
		ExchangeRate:                block.GetExchRate(),
		BlockHeight:                 block.GetDBHeight(),
		TransactionCount:            uint32(len(block.GetTransactions())),
		Transactions:                mapTransactions(block.GetTransactions(), block.GetEndOfPeriod()),
	}
	return result
}

func mapTransactions(transactions []interfaces.ITransaction, endOfPeriod [10]int) []*eventmessages.Transaction {
	result := make([]*eventmessages.Transaction, 0)
	for i, transaction := range transactions {
		err := transaction.ValidateSignatures()
		if err == nil {
			// The endOfPeriod array contains the transaction count at the new minute transition time
			// The minuteMark will contain the value of during which minute the transaction took place which is one higher than the position within the array
			var minuteMark = 0
			for ; minuteMark < len(endOfPeriod) && endOfPeriod[minuteMark] > 0; minuteMark++ {
				if endOfPeriod[minuteMark] >= i {
					minuteMark++
					break
				}
			}
			result = append(result, mapTransaction(transaction, minuteMark))
		}
	}
	return result
}

func mapTransaction(transaction interfaces.ITransaction, minuteNumber int) *eventmessages.Transaction {
	result := &eventmessages.Transaction{
		TransactionID:                 transaction.GetSigHash().Bytes(),
		BlockHeight:                   transaction.GetBlockHeight(),
		MinuteNumber:                  uint32(minuteNumber),
		Timestamp:                     ConvertTimeToTimestamp(transaction.GetTimestamp().GetTime()),
		FactoidInputs:                 mapTransactionAddresses(transaction.GetInputs()),
		FactoidOutputs:                mapTransactionAddresses(transaction.GetOutputs()),
		EntryCreditOutputs:            mapTransactionAddresses(transaction.GetECOutputs()),
		RedeemConditionDataStructures: mapRCDs(transaction.GetRCDs()),
		SignatureBlocks:               mapSignatureBlocks(transaction.GetSignatureBlocks()),
	}
	return result
}

func mapTransactionAddresses(inputs []interfaces.ITransAddress) []*eventmessages.TransactionAddress {
	result := make([]*eventmessages.TransactionAddress, len(inputs))
	for i, input := range inputs {
		result[i] = mapTransactionAddress(input)
	}
	return result
}

func mapTransactionAddress(address interfaces.ITransAddress) *eventmessages.TransactionAddress {
	result := &eventmessages.TransactionAddress{
		Amount:  address.GetAmount(),
		Address: address.GetAddress().Bytes(),
	}
	return result
}

func mapRCDs(rcds []interfaces.IRCD) []*eventmessages.RCD {
	result := make([]*eventmessages.RCD, len(rcds))
	for i, rcd := range rcds {
		result[i] = mapRCD(rcd)
	}
	return result
}

type RCD interface {
	GetPublicKey() []byte
}

func mapRCD(rcd interfaces.IRCD) *eventmessages.RCD {
	result := &eventmessages.RCD{}

	if rcd1, ok := rcd.(RCD); ok {
		result.Rcd = &eventmessages.RCD_Rcd1{
			Rcd1: &eventmessages.RCD1{
				PublicKey: rcd1.GetPublicKey(),
			},
		}
	}
	return result
}

func mapSignatureBlocks(blocks []interfaces.ISignatureBlock) []*eventmessages.FactoidSignatureBlock {
	result := make([]*eventmessages.FactoidSignatureBlock, len(blocks))
	for i, block := range blocks {
		result[i] = mapSignatureBlock(block)
	}
	return result
}

func mapSignatureBlock(block interfaces.ISignatureBlock) *eventmessages.FactoidSignatureBlock {
	result := &eventmessages.FactoidSignatureBlock{
		Signature: mapFactoidSignatureBlockSignatures(block.GetSignatures()),
	}
	return result
}

func mapFactoidSignatureBlockSignatures(signatures []interfaces.ISignature) [][]byte {
	result := make([][]byte, len(signatures))
	for i, signature := range signatures {
		result[i] = signature.Bytes()
	}
	return result
}
