// +build longtest

package engine_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"

	"github.com/FactomProject/factomd/activations"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestDBsigElectionEvery2Block_long(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	iterations := 1
	state := SetupSim("LLLLLLAF", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 35, 6, 6, t)

	RunCmd("S10") // Set Drop Rate to 1.0 on everyone

	for j := 0; j < iterations; j++ {
		// for leader 1 thu 7 kill each in turn
		for i := 1; i < 7; i++ {
			s := GetFnodes()[i].State
			if !s.IsLeader() {
				panic("Can't kill a audit and cause an election")
			}
			WaitForMinute(s, 9) // wait till the victim is at minute 9
			// wait till minute flips
			for s.CurrentMinute != 0 {
				runtime.Gosched()
			}
			s.SetNetStateOff(true) // kill the victim
			s.LogPrintf("faulting", "Stopped %s\n", s.FactomNodeName)
			WaitForMinute(state, 1) // Wait till FNode0 move ahead a minute (the election is over)
			s.LogPrintf("faulting", "Start %s\n", s.FactomNodeName)
			s.SetNetStateOff(false) // resurrect the victim

			// Wait till the should have updated by DBSTATE
			WaitBlocks(state, 2)
			WaitForMinute(state, 1)
			WaitForAllNodes(state)

			CheckAuthoritySet(t) // check the authority set is as expected
		}
	}
	WaitForAllNodes(state)
	ShutDownEverything(t)

}

func TestGrants_long(t *testing.T) {
	if RanSimTest {
		return
	}

	makeExpected := func(grants []state.HardGrant) []interfaces.ITransAddress {
		var rval []interfaces.ITransAddress
		for _, g := range grants {
			rval = append(rval, factoid.NewOutAddress(g.Address, g.Amount))
		}
		return rval
	}

	RanSimTest = true

	state0 := SetupSim("LAF", map[string]string{"--debuglog": "", "--faulttimeout": "10", "--blktime": "5"}, 300, 0, 0, t)

	grants := state.GetHardCodedGrants()

	// find all the heights we care about
	heights := map[uint32][]state.HardGrant{}
	min := uint32(9999999)
	max := uint32(0)
	grantBalances := map[string]int64{} // Compute the expected final balances
	// TODO: (does not account for cancels)
	for _, g := range grants {
		heights[g.DBh] = append(heights[g.DBh], g)
		if min > g.DBh {
			min = g.DBh
		}
		if max < g.DBh {
			max = g.DBh
		}
		// keep a list of grant addresses
	}

	// Build a list of grant addresses
	for _, g := range grants {
		userAddr := primitives.ConvertFctAddressToUserStr(g.Address)
		_, ok := grantBalances[userAddr]
		if !ok {
			grantBalances[userAddr] = state0.FactoidState.GetFactoidBalance(g.Address.Fixed()) // Save initial balance
		}
		grantBalances[userAddr] += int64(g.Amount) // Add the grant amount
	}

	fmt.Println("Waiting for grant payout")
	// run the state till we are past the 100 block delay and check the final balances
	WaitBlocks(state0, int(max+1+constants.COINBASE_DECLARATION+constants.COINBASE_PAYOUT_FREQUENCY*2))

	// check the final balances of the accounts
	for addr, balance := range grantBalances {
		factoidBalance := state0.FactoidState.GetFactoidBalance(factoid.NewAddress(primitives.ConvertUserStrToAddress(addr)).Fixed())
		if balance != factoidBalance {
			t.Errorf("FinalBalanceMismatch for %s. Got %d expected %d", addr, balance, factoidBalance)
		}
	}

	// loop thru the dbheights  to get the admin block and check them and make sure the payouts get returned
	for dbheight := uint32(min - constants.COINBASE_PAYOUT_FREQUENCY*2); dbheight <= uint32(max+constants.COINBASE_PAYOUT_FREQUENCY*2); dbheight++ {
		expected := makeExpected(heights[dbheight])
		gotGrants := state.GetGrantPayoutsFor(dbheight)
		if len(expected) != len(gotGrants) {
			t.Errorf("Expected %d grants but found %d", len(expected), len(gotGrants))
		} else if len(expected) > 0 {
			fmt.Printf("Got %d expected grants at %d\n", len(expected), dbheight)
		}

		for i, _ := range expected {
			if !expected[i].GetAddress().IsSameAs(gotGrants[i].GetAddress()) ||
				expected[i].GetAmount() != gotGrants[i].GetAmount() ||
				expected[i].GetUserAddress() != gotGrants[i].GetUserAddress() {
				t.Errorf("Expected: %v ", expected[i])
				t.Errorf("but found %v for grant #%d at %d", gotGrants[i], i, dbheight)
			} else {
				fmt.Printf("Got grants %v\n", expected[i])
			}
			//fmt.Println(p.GetAmount(), p.GetUserAddress())
		}
		//descriptorHeight := dbheight - constants.COINBASE_DECLARATION

		ablock, err := state0.DB.FetchABlockByHeight(dbheight)
		if err != nil {
			panic(fmt.Sprintf("Missing coinbase, admin block at height %d could not be retrieved", dbheight))
		}

		abe := ablock.FetchCoinbaseDescriptor()
		if abe != nil {
			desc := abe.(*adminBlock.CoinbaseDescriptor)
			coinBaseOutputs := map[string]uint64{}
			for _, o := range desc.Outputs {
				coinBaseOutputs[primitives.ConvertFctAddressToUserStr(o.GetAddress())] = o.GetAmount()
			}
			if len(expected) != len(coinBaseOutputs) && !(len(coinBaseOutputs) == 1 && dbheight%constants.COINBASE_PAYOUT_FREQUENCY == 0) {
				t.Errorf("Expected %d grants but found %d at height %d", len(expected), len(coinBaseOutputs), dbheight)
				PrintList("coinbase", coinBaseOutputs)
			}
			for i, _ := range expected {
				address := expected[i].GetUserAddress()
				cbAmount := coinBaseOutputs[address]
				amount := expected[i].GetAmount()
				if amount != cbAmount {
					t.Errorf("Expected: %v ", expected[i])
					t.Errorf("but found %v:%v for grant #%d at %d", address, cbAmount, i, dbheight)
				}
				//fmt.Println(p.GetAmount(), p.GetUserAddress())
			}
		}
	} // for all dbheights {...}

	WaitForAllNodes(state0)

	ShutDownEverything(t)
}

func TestTestNetCoinBaseActivation_long(t *testing.T) {
	if RanSimTest {
		return
	}

	state0 := SetupSim("LAF", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 168, 0, 0, t)
	fmt.Println("Simulation configured")
	nextBlock := uint32(11 + constants.COINBASE_DECLARATION) // first grant is at 11 so it pays at 21
	fmt.Println("Wait till first grant should payout")
	WaitForBlock(state0, int(nextBlock)) // wait for the first coin base payout to be generated
	factoidState0 := state0.FactoidState.(*state.FactoidState)
	CBT := factoidState0.GetCoinbaseTransaction(nextBlock, state0.GetLeaderTimestamp())
	oldCBDelay := constants.COINBASE_DECLARATION
	if oldCBDelay != 10 {
		t.Fatalf("constants.COINBASE_DECLARATION = %d expect 10\n", constants.COINBASE_DECLARATION)
	}
	if len(CBT.GetOutputs()) != 1 {
		t.Fatalf("Expected first payout at block %d\n", nextBlock)
	} else {
		fmt.Println("Got first payout")
	}

	fmt.Println("Wait till activation height")
	blk := activations.ActivationMap[activations.TESTNET_COINBASE_PERIOD].ActivationHeight["LOCAL"]
	WaitForBlock(state0, blk)
	if constants.COINBASE_DECLARATION != 140 {
		t.Fatalf("constants.COINBASE_DECLARATION = %d expect 140\n", constants.COINBASE_DECLARATION)
	}

	nextBlock += constants.COINBASE_DECLARATION - oldCBDelay + 1
	fmt.Println("Wait till second grant should payout with the new activation height")
	WaitForBlock(state0, int(nextBlock+1)) // next payout passed new activation (should be paid)
	CBT = factoidState0.GetCoinbaseTransaction(nextBlock, state0.GetLeaderTimestamp())
	if len(CBT.GetOutputs()) != 0 {
		t.Fatalf("Expected first payout at block %d\n", nextBlock)
	}
	fmt.Println("Wait to shut down")
	StatusEveryMinute(state0)
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}
