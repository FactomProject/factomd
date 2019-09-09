package eventservices

import (
	"github.com/FactomProject/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
)

func mapFactoidBlock(block interfaces.IFBlock) *eventmessages.FactoidBlock {
	result := &eventmessages.FactoidBlock{
		BodyMerkleRoot: &eventmessages.Hash{
			HashValue: block.GetBodyMR().Bytes(),
		},
		PreviousKeyMerkleRoot: &eventmessages.Hash{
			HashValue: block.GetPrevKeyMR().Bytes(),
		},
		PreviousLedgerKeyMerkleRoot: &eventmessages.Hash{
			HashValue: block.GetLedgerKeyMR().Bytes(),
		},
		ExchRate:     block.GetExchRate(),
		BlockHeight:  block.GetDBHeight(),
		Transactions: mapTransactions(block.GetTransactions()),
	}
	return result
}

func mapTransactions(transactions []interfaces.ITransaction) []*eventmessages.Transaction {
	result := make([]*eventmessages.Transaction, 0)
	for _, transaction := range transactions {
		err := transaction.ValidateSignatures()
		if err == nil {
			result = append(result, mapTransaction(transaction))
		}
	}
	return result
}

func mapTransaction(transaction interfaces.ITransaction) *eventmessages.Transaction {
	result := &eventmessages.Transaction{
		TransactionId: &eventmessages.Hash{
			HashValue: transaction.GetSigHash().Bytes(),
		},
		BlockHeight:        transaction.GetBlockHeight(),
		Timestamp:          convertTimeToTimestamp(transaction.GetTimestamp().GetTime()),
		Inputs:             mapTransactionAddresses(transaction.GetInputs()),
		Outputs:            mapTransactionAddresses(transaction.GetOutputs()),
		OutputEntryCredits: mapTransactionAddresses(transaction.GetECOutputs()),
		Rcds:               mapRCDs(transaction.GetRCDs()),
		SignatureBlocks:    mapSignatureBlocks(transaction.GetSignatureBlocks()),
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
		Amount: address.GetAmount(),
		Address: &eventmessages.Hash{
			HashValue: address.GetAddress().Bytes(),
		},
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

func mapRCD(rcd interfaces.IRCD) *eventmessages.RCD {
	result := &eventmessages.RCD{}
	switch rcd.(type) {
	case factoid.IRCD_1:
		rcd1 := rcd.(factoid.IRCD_1)
		result.Value = &eventmessages.RCD_Rcd1{
			Rcd1: &eventmessages.RCD1{
				PublicKey: rcd1.GetPublicKey(),
			},
		}

	case factoid.IRCD:
		/*		rcd2 := rcd.(factoid.IRCD)  TODO rcd2 is not implemented?
				evRcd2 := eventmessages.RCD2{
					M:          rcd.M,
					N:          rcd.NumberOfSignatures(),
					NAddresses: rcd2.GetHash().Bytes(),
				}
				evRcd2Value := &eventmessages.RCD_Rcd2{Rcd2: evRcd2}
				result.Value = evRcd2Value*/
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
		Signature: mapFactoidSignatureBlockSingatures(block.GetSignatures()),
	}
	return result
}

func mapFactoidSignatureBlockSingatures(signatures []interfaces.ISignature) []*eventmessages.FactoidSignature {
	result := make([]*eventmessages.FactoidSignature, len(signatures))
	for i, signature := range signatures {
		result[i] = &eventmessages.FactoidSignature{
			SignatureValue: signature.Bytes(),
		}
	}
	return result
}
