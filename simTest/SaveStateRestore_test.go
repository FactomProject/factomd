package simtest

import (
	"fmt"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFastBootSaveAndRestore(t *testing.T) {
	var saveRate = 4
	var state0 *state.State
	var bankSecret string = "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"
	var depositAddresses []string
	var numAddresses = 1
	var depositCount int64 = 0
	var ecPrice uint64 = 10000

	for i := 0; i < numAddresses; i++ {
		_, addr := engine.RandomFctAddressPair()
		depositAddresses = append(depositAddresses, addr)
	}


	mkTransactions := func() {
		depositCount++
		for i := range depositAddresses {
			fmt.Printf("TXN %v %v => %v \n", depositCount, depositAddresses[i], depositAddresses[i])
			time.Sleep(time.Millisecond*90)
			engine.SendTxn(state0, 1, bankSecret, depositAddresses[i], ecPrice)
		}
	}

	startSim := func(nodes string, maxHeight int) {
		RanSimTest = true
		state0 = SetupSim(
			nodes,
			map[string]string{"--debuglog": ".", "--fastsaverate": fmt.Sprintf("%v", saveRate)},
			maxHeight,
			0,
			0,
			t,
		)
	}

	stopSim := func() {
		WaitForAllNodes(state0)
		ShutDownEverything(t)
		state0 = nil
	}

	abortSim := func(msg string) {
		ShutDownEverything(t)
		println("ABORT: "+msg)
		t.Fatal(msg)
	}

	newState := func() (*state.State, int){
		index := len(engine.GetFnodes())
		newState := state0.Clone(index)
		return newState.(*state.State), index
	}
	t.Run("run sim to create fastboot", func(t *testing.T) {
		if RanSimTest {
			return
		}

		startSim("LF", 25)

		t.Run("add transactions to fastboot block", func(t *testing.T) {
			mkTransactions()
			WaitForBlock(state0, 5) // REVIEW: maybe wait another fastboot period?
			mkTransactions()
			WaitForBlock(state0, 6)
		})

		targetState := state0
		WaitForBlock(targetState, saveRate*2+2)

		snapshot, _ := targetState.GetMapDB().Clone()

		assert.NotNil(t, targetState.StateSaverStruct.TmpState)
		mkTransactions()

		t.Run("create fnode02 without fastboot", func(t *testing.T) {
			s, _ := newState()
			_, i := AddNode(s)
			engine.StartFnode(i, true)
		})

		t.Run("create fnode03 with fastboot", func(t *testing.T) {

			// transplant database to new node
			s, _ := newState()
			node, i := AddNode(s)
			_ = snapshot
			s.SetMapDB(targetState.GetMapDB()) // KLUDGE using same DB

			t.Run("restore state from fastboot", func(t *testing.T) {

				// restore savestate
				err := s.StateSaverStruct.LoadDBStateListFromBin(s.DBStates, targetState.StateSaverStruct.TmpState)
				assert.Nil(t, err)

				assert.False(t, s.IsLeader())
				assert.True(t, s.DBHeightAtBoot > 0, "Failed to restore db height on fnode03")

				if s.DBHeightAtBoot == 0 {
					abortSim("Fastboot was not restored properly")
				} else {
					fmt.Printf("RESTORED DBHeight: %v\n", s.DBHeightAtBoot)
				}

			})

			engine.StartFnode(i, true)
			assert.True(t, node.Running)

			WaitBlocks(state0, 1)

			if ! node.State.DBFinished {
				abortSim("DBFinished is not set")
			}

			WaitBlocks(state0, 3)

			if len(node.State.Holding) > 40 {
				abortSim("holding queue is backed up")
			}
		})

		t.Run("compare permanent balances on each node", func(t *testing.T) {
			stopSim() // graceful stop
			for i, node := range engine.GetFnodes() {
				for _, addr := range depositAddresses {
					bal := engine.GetBalance(node.State, addr)
					msg := fmt.Sprintf("CHKBAL Node%v %v => balance: %v expect: %v \n", i, addr, bal, depositCount)
					println(msg)
					assert.Equal(t, depositCount, bal, msg)
				}
			}
		})

	})
}
