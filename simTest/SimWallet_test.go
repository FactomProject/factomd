package simtest

import (
	"fmt"
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/FactomProject/factom"
)

func TestSimWallet(t *testing.T) {

	a := simulation.AccountFromFctSecret("Fs31kMMKBSCDwa7tSEzjQ4EfGeXARdUS22oBEJKNSJWbLAMTsEHr")
	b := simulation.AccountFromFctSecret("Fs2BNvoDgSoGJpWg4PvRUxqvLE28CQexp5FZM9X5qU6QvzFBUn6D")

	fmt.Printf("A: %s", a)
	fmt.Printf("B: %s", b)

	state0 := simulation.SetupSim("L", map[string]string{"--debuglog": ""}, 80, 0, 0, t)

	var oneFct uint64 = factom.FactoidToFactoshi("1")
	a.FundFCT(10 * oneFct)                                                 // fund fct from coinbase 'bank'
	a.SendFCT(b, oneFct)                                                   // fund Account b from Acct a
	simulation.WaitForAnyFctBalance(state0, a.FctPub())                    // wait for non-zero
	simulation.WaitForFctBalanceUnder(state0, a.FctPub(), int64(9*oneFct)) // wait for at least 1 fct
	simulation.WaitForFctBalanceOver(state0, b.FctPub(), int64(oneFct-1))  // wait for at least 1 fct

	a.FundEC(10)                                          // fund EC from coinbase 'bank'
	simulation.WaitForEcBalanceOver(state0, a.EcPub(), 1) // wait for at least 1 ec

	b.ConvertEC(10) // fund EC from account b
	simulation.WaitForAnyEcBalance(state0, b.EcPub())
	simulation.WaitForEcBalanceUnder(state0, b.EcPub(), 11)

	simulation.WaitBlocks(state0, 1)
	simulation.ShutDownEverything(t)

}
