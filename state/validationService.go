package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
)

func NewValidationService() chan ValidationMsg {
	c := make(chan ValidationMsg)
	go ValidationServiceLoop(c)
	return c
}

const (
	MessageTypeGetFactoidBalance = iota
	MessageTypeGetECBalance
	MessageTypeUpdateTransaction
	MessageTypeGetFactoshisPerEC
	MessageTypeSetFactoshisPerEC
	MessageTypeResetBalances
	MessageTypeValidate
)

type ValidationMsg struct {
	MessageType    int
	Address        [32]byte
	Transaction    interfaces.ITransaction
	ECTransaction  interfaces.IECBlockEntry
	FactoshisPerEC uint64

	ReturnChannel chan ValidationResponseMsg
}

type ValidationResponseMsg struct {
	Error          error
	Balance        int64
	FactoshisPerEC uint64
}

func ValidationServiceLoop(input chan ValidationMsg) {
	type ValidationState struct {
		NumTransactions int
		FactoidBalances map[[32]byte]int64
		ECBalances      map[[32]byte]int64
		FactoshisPerEC  uint64
	}

	vs := new(ValidationState)
	vs.FactoshisPerEC = 1
	vs.FactoidBalances = map[[32]byte]int64{}
	vs.ECBalances = map[[32]byte]int64{}

	for {
		msg := <-input
		switch msg.MessageType {
		case MessageTypeGetFactoidBalance:
			v := vs.FactoidBalances[msg.Address]
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				resp.Balance = v
				msg.ReturnChannel <- resp
			}
			break

		case MessageTypeGetECBalance:
			v := vs.ECBalances[msg.Address]
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				resp.Balance = v
				msg.ReturnChannel <- resp
			}
			break

		case MessageTypeUpdateTransaction:

			if msg.Transaction == nil && msg.ECTransaction == nil {
				if msg.ReturnChannel != nil {
					var resp ValidationResponseMsg
					resp.Error = fmt.Errorf("No transaction provided")
					msg.ReturnChannel <- resp
				}
				break
			}

			if msg.Transaction != nil {
				trans := msg.Transaction
				for _, input := range trans.GetInputs() {
					vs.FactoidBalances[input.GetAddress().Fixed()] = vs.FactoidBalances[input.GetAddress().Fixed()] - int64(input.GetAmount())
				}
				for _, output := range trans.GetOutputs() {
					vs.FactoidBalances[output.GetAddress().Fixed()] = vs.FactoidBalances[output.GetAddress().Fixed()] + int64(output.GetAmount())
				}
				for _, ecOut := range trans.GetECOutputs() {
					ecbal := int64(ecOut.GetAmount()) / int64(vs.FactoshisPerEC)
					vs.ECBalances[ecOut.GetAddress().Fixed()] = vs.ECBalances[ecOut.GetAddress().Fixed()] + ecbal
				}
				vs.NumTransactions++
				if msg.ReturnChannel != nil {
					var resp ValidationResponseMsg
					msg.ReturnChannel <- resp
				}
			}

			if msg.ECTransaction != nil {
				trans := msg.ECTransaction
				var resp ValidationResponseMsg

				switch trans.ECID() {
				case entryCreditBlock.ECIDServerIndexNumber:
					resp.Error = fmt.Errorf("Invalid transaction provided")
					msg.ReturnChannel <- resp
					break

				case entryCreditBlock.ECIDMinuteNumber:
					resp.Error = fmt.Errorf("Invalid transaction provided")
					msg.ReturnChannel <- resp
					break

				case entryCreditBlock.ECIDChainCommit:
					t := trans.(*entryCreditBlock.CommitChain)
					vs.ECBalances[t.ECPubKey.Fixed()] = vs.ECBalances[t.ECPubKey.Fixed()] - int64(t.Credits)
					vs.NumTransactions++

					msg.ReturnChannel <- resp
					break

				case entryCreditBlock.ECIDEntryCommit:
					t := trans.(*entryCreditBlock.CommitEntry)
					vs.ECBalances[t.ECPubKey.Fixed()] = vs.ECBalances[t.ECPubKey.Fixed()] - int64(t.Credits)
					vs.NumTransactions++

					msg.ReturnChannel <- resp
					break

				case entryCreditBlock.ECIDBalanceIncrease:
					t := trans.(*entryCreditBlock.IncreaseBalance)
					vs.ECBalances[t.ECPubKey.Fixed()] = vs.ECBalances[t.ECPubKey.Fixed()] + int64(t.NumEC)
					vs.NumTransactions++

					msg.ReturnChannel <- resp
					break

				default:
					resp.Error = fmt.Errorf("Unknown EC transaction provided")
					msg.ReturnChannel <- resp
					break
				}
			}

			break

		case MessageTypeResetBalances:
			vs.FactoidBalances = map[[32]byte]int64{}
			vs.ECBalances = map[[32]byte]int64{}
			vs.NumTransactions = 0
			break

		case MessageTypeGetFactoshisPerEC:
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				resp.FactoshisPerEC = vs.FactoshisPerEC
				msg.ReturnChannel <- resp
			}
			break

		case MessageTypeSetFactoshisPerEC:
			vs.FactoshisPerEC = msg.FactoshisPerEC
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				msg.ReturnChannel <- resp
			}
			break

		case MessageTypeValidate:
			var resp ValidationResponseMsg
			var sums = make(map[[32]byte]uint64, 10) // Look at the sum of an address's inputs
			trans := msg.Transaction
			for _, input := range trans.GetInputs() { //    to a transaction.
				bal, err := factoid.ValidateAmounts(sums[input.GetAddress().Fixed()], input.GetAmount())
				if err != nil {
					if msg.ReturnChannel != nil {
						resp.Error = err
						msg.ReturnChannel <- resp
					}
					break
				}
				if int64(bal) > vs.FactoidBalances[input.GetAddress().Fixed()] {
					if msg.ReturnChannel != nil {
						resp.Error = fmt.Errorf("Not enough funds in input addresses for the transaction")
						msg.ReturnChannel <- resp
					}
					break
				}
				sums[input.GetAddress().Fixed()] = bal
			}
			msg.ReturnChannel <- resp
			break

		default:
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				resp.Error = fmt.Errorf("Unknown MessageType")
				msg.ReturnChannel <- resp
			}
		}
	}
}

/*
	}*/
