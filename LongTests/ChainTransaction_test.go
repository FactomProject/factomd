package longtests

import (
	"fmt"
	"github.com/FactomProject/factom"
	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
	"time"
)

func TestChainedTransactions(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	// a genesis block address w/ funding
	bankSecret := "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"
	bankAddress := "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q"

	var depositSecrets []string
	var depositAddresses []string

	for i:=0; i<120; i++  {
		priv, addr := RandomFctAddressPair()
		depositSecrets = append(depositSecrets, priv)
		depositAddresses = append(depositAddresses, addr)
	}

	var maxBlocks = 500
	state0 := SetupSim("LAF", map[string]string{"--debuglog": ".",}, maxBlocks+1, 0, 0, t)
	var ecPrice uint64 = state0.GetFactoshisPerEC() //10000
	var oneFct uint64 = factom.FactoidToFactoshi("1")

	waitForDeposit := func(i int, amt uint64) uint64 {
		balance := GetBalance(state0, depositAddresses[i])
		TimeNow(state0)
		fmt.Printf("%v waitForDeposit %v %v - %v = diff: %v \n", i, depositAddresses[i], balance, amt, balance-int64(amt))
		var waited bool
		for balance != int64(amt) {
			waited = true
			balance = GetBalance(state0, depositAddresses[i])
			time.Sleep(time.Millisecond*100)
		}
		if waited {
			fmt.Printf("%v waitForDeposit %v %v - %v = diff: %v \n", i, depositAddresses[i], balance, amt, balance-int64(amt))
			TimeNow(state0)
		}
		return uint64(balance)
	}
	_ = waitForDeposit

	initialBalance := 10*oneFct
	fee := 12*ecPrice

	prepareTransactions := func(bal uint64) ([]func(), uint64, int) {

		var transactions []func()
		var i int

		for i = 0; i < len(depositAddresses)-1; i += 1 {
			bal -= fee

			in := i
			out := i+1
			send := bal

			txn := func() {
				//fmt.Printf("TXN %v %v => %v \n", send, depositAddresses[in], depositAddresses[out])
				SendTxn(state0, send, depositSecrets[in], depositAddresses[out], ecPrice)
			}
			transactions = append(transactions, txn)
		}
		return transactions, bal, i
	}

	// offset to send initial blocking transaction
	offset := 1

	mkTransactions := func() { // txnGenerator
		// fund the start address
		SendTxn(state0, initialBalance, bankSecret, depositAddresses[0], ecPrice)
		WaitMinutes(state0, 1)
		waitForDeposit(0, initialBalance)
		transactions, finalBalance, finalAddress := prepareTransactions(initialBalance)

		var sent []int
		var unblocked bool = false

		for i:=1; i<len(transactions); i++ {
			sent = append(sent, i)
			//fmt.Printf("offset: %v <=> i:%v", offset, i)
			if i == offset {
				fmt.Printf("\n==>TXN offset%v\n", offset)
				transactions[0]() // unblock the transactions
				unblocked = true
			}
			transactions[i]()
		}
		if ! unblocked{
			transactions[0]() // unblock the transactions
		}
		offset++ // next time start further in the future
		fmt.Printf("send chained transations")
		waitForDeposit(finalAddress, finalBalance)

		// empty final address returning remaining funds to bank
		SendTxn(state0, finalBalance-fee, depositSecrets[finalAddress], bankAddress, ecPrice)
		waitForDeposit(finalAddress, 0)
	}
	_ = mkTransactions

	for x:= 1; x<= 120; x++ {
		mkTransactions()
		WaitBlocks(state0, 1)
	}

	WaitForAllNodes(state0)
	ShutDownEverything(t)
}
