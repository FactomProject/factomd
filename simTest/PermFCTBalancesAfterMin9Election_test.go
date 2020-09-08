package simtest

import (
	"fmt"
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
	"time"

	"github.com/FactomProject/factomd/fnode"

	"github.com/stretchr/testify/assert"
)

var bankSecret string = "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"
var depositAddresses []string

// generate addresses & private keys
func createDepositAddresses() {
	for i := 0; i < 1; i++ {
		_, addr := simulation.RandomFctAddressPair()
		depositAddresses = append(depositAddresses, addr)
	}
}

func TestPermFCTBalancesAfterMin9Election(t *testing.T) {
	createDepositAddresses()
	state0 := simulation.SetupSim("LLAL", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 10, 1, 1, t)

	var depositCount int64 = 0
	var ecPrice uint64 = state0.GetFactoshisPerEC() //10000

	mkTransactions := func() {
		depositCount++
		for i := range depositAddresses {
			fmt.Printf("TXN %v %v => %v \n", depositCount, depositAddresses[i], depositAddresses[i])
			time.Sleep(time.Millisecond * 90)
			simulation.SendTxn(state0, 1, bankSecret, depositAddresses[i], ecPrice)
		}
	}

	simulation.StatusEveryMinute(state0)
	simulation.CheckAuthoritySet(t)

	state3 := fnode.Get(3).State
	if !state3.IsLeader() {
		panic("Can't kill a audit and cause an election")
	}
	simulation.RunCmd("3")
	simulation.WaitForMinute(state3, 9) // wait till the victim is at minute 9
	simulation.RunCmd("x")
	go mkTransactions()
	simulation.WaitMinutes(state0, 1) // Wait till fault completes
	simulation.RunCmd("x")
	simulation.WaitBlocks(state0, 2)    // wait till the victim is back as the audit server
	simulation.WaitForMinute(state0, 1) // Wait till ablock is loaded
	simulation.WaitForAllNodes(state0)
	simulation.WaitForMinute(state3, 1) // Wait till node 3 is following by minutes
	//WaitBlocks(state3, 2)

	simulation.WaitForAllNodes(state0)
	simulation.ShutDownEverything(t)

	for i, node := range fnode.GetFnodes() {
		for _, addr := range depositAddresses {
			bal := simulation.GetBalance(node.State, addr)
			msg := fmt.Sprintf("Node%v %v => balance: %v expected: %v \n", i, addr, bal, depositCount)
			assert.Equal(t, depositCount, bal, msg)
		}
	}

}
