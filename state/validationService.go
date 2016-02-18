package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"runtime/debug"
)

var _ = fmt.Println
var _ = debug.PrintStack

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
	MessageTypeClearRealTime
	MessageTypeValidate
)

type ValidationMsg struct {
	// If realtime, the updated balance is updated in the Temp space.
	// If not realtime (i.e. processing a block), balance is updated in the 
	// permanent space.
	Realtime	   bool     
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

type ValidationState struct {
	NumTransactions int
	
	// Permanent balances from processing blocks. 
	FactoidBalancesP map[[32]byte]int64
	ECBalancesP      map[[32]byte]int64
	
	// Temporary balances from updating transactions in real time.
	FactoidBalancesT map[[32]byte]int64
	ECBalancesT      map[[32]byte]int64
	
	FactoshisPerEC  uint64
}

func (vs *ValidationState) GetF(adr [32]byte) int64 {
	if v,ok := vs.FactoidBalancesT[adr]; !ok {
		v = vs.FactoidBalancesP[adr]
		return v
	}else{
		return v
	}
}

func (vs *ValidationState) PutF(rt bool, adr [32]byte, v int64) {
	if rt {
		vs.FactoidBalancesT[adr] = v
	}else{
		vs.FactoidBalancesP[adr] = v
	}
}

func (vs *ValidationState) GetE(adr [32]byte) int64 {
	if v,ok := vs.ECBalancesT[adr]; !ok {
		v = vs.ECBalancesP[adr]
		return v
	}else{
		return v
	}
}

func (vs *ValidationState) PutE(rt bool, adr [32]byte, v int64) {
	if rt {
		vs.ECBalancesT[adr] = v
	}else{
		vs.ECBalancesP[adr] = v
	}
}


func ValidationServiceLoop(input chan ValidationMsg) {


	vs := new(ValidationState)
	vs.FactoshisPerEC = 1
	vs.FactoidBalancesP = map[[32]byte]int64{}
	vs.ECBalancesP = map[[32]byte]int64{}
	vs.FactoidBalancesT = map[[32]byte]int64{}
	vs.ECBalancesT = map[[32]byte]int64{}
	
	for {
		msg := <-input
		switch msg.MessageType {
			
		case MessageTypeClearRealTime:
			vs.FactoidBalancesT = map[[32]byte]int64{}
			vs.ECBalancesT = map[[32]byte]int64{}
			
		case MessageTypeGetFactoidBalance:
			v := vs.GetF(msg.Address)
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				resp.Balance = v
				msg.ReturnChannel <- resp
			}
		
		case MessageTypeGetECBalance:
			v := vs.GetE(msg.Address)
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				resp.Balance = v
				msg.ReturnChannel <- resp
			}

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
				rt := msg.Realtime
				for _, input := range trans.GetInputs() {
					vs.PutF(rt,input.GetAddress().Fixed(), vs.GetF(input.GetAddress().Fixed()) - int64(input.GetAmount()))
				}
				for _, output := range trans.GetOutputs() {
					vs.PutF(rt,output.GetAddress().Fixed(), vs.GetF(output.GetAddress().Fixed()) + int64(output.GetAmount()))
				}
				for _, ecOut := range trans.GetECOutputs() {
					ecbal := int64(ecOut.GetAmount()) / int64(vs.FactoshisPerEC)
					vs.PutE(rt,ecOut.GetAddress().Fixed(), vs.GetE(ecOut.GetAddress().Fixed()) + ecbal)
				}
				vs.NumTransactions++
				if msg.ReturnChannel != nil {
					var resp ValidationResponseMsg
					msg.ReturnChannel <- resp
				}
			}

			if msg.ECTransaction != nil {
				rt := msg.Realtime
				trans := msg.ECTransaction
				var resp ValidationResponseMsg

				switch trans.ECID() {
				case entryCreditBlock.ECIDServerIndexNumber:
					resp.Error = fmt.Errorf("Invalid transaction provided")
					msg.ReturnChannel <- resp

				case entryCreditBlock.ECIDMinuteNumber:
					resp.Error = fmt.Errorf("Invalid transaction provided")
					msg.ReturnChannel <- resp

				case entryCreditBlock.ECIDChainCommit:
					t := trans.(*entryCreditBlock.CommitChain)
					vs.PutE(rt,t.ECPubKey.Fixed(), vs.GetE(t.ECPubKey.Fixed()) - int64(t.Credits))
					vs.NumTransactions++
					msg.ReturnChannel <- resp

				case entryCreditBlock.ECIDEntryCommit:
					t := trans.(*entryCreditBlock.CommitEntry)
					vs.PutE(rt,t.ECPubKey.Fixed(), vs.GetE(t.ECPubKey.Fixed()) - int64(t.Credits))
					vs.NumTransactions++

					msg.ReturnChannel <- resp

				case entryCreditBlock.ECIDBalanceIncrease:
					t := trans.(*entryCreditBlock.IncreaseBalance)
					vs.PutE(rt,t.ECPubKey.Fixed(), vs.GetE(t.ECPubKey.Fixed()) + int64(t.NumEC))
					vs.NumTransactions++

					msg.ReturnChannel <- resp

				default:
					resp.Error = fmt.Errorf("Unknown EC transaction provided")
					msg.ReturnChannel <- resp
				}
			}

		case MessageTypeResetBalances:
			vs.FactoidBalancesP = map[[32]byte]int64{}
			vs.ECBalancesP = map[[32]byte]int64{}
			vs.FactoidBalancesT = map[[32]byte]int64{}
			vs.ECBalancesT = map[[32]byte]int64{}
			vs.NumTransactions = 0

		case MessageTypeGetFactoshisPerEC:
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				resp.FactoshisPerEC = vs.FactoshisPerEC
				msg.ReturnChannel <- resp
			}

		case MessageTypeSetFactoshisPerEC:
			vs.FactoshisPerEC = msg.FactoshisPerEC
			if msg.ReturnChannel != nil {
				var resp ValidationResponseMsg
				msg.ReturnChannel <- resp
			}

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
				if int64(bal) > vs.GetF(input.GetAddress().Fixed()) {
					if msg.ReturnChannel != nil {
						resp.Error = fmt.Errorf("Not enough funds in input addresses for the transaction")
						msg.ReturnChannel <- resp
					}
					break
				}
				sums[input.GetAddress().Fixed()] = bal
			}
			msg.ReturnChannel <- resp

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
