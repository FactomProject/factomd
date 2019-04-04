package simtest

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
)

func TestSendingCommitAndReveal(t *testing.T) {
	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	logName := "simTest"
	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	extids := [][]byte{encode("foo"), encode("bar")}
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := GetBankAccount()

	t.Run("generate accounts", func(t *testing.T) {
		println(b.String())
		println(a.String())
	})

	t.Run("Run sim to create entries", func(t *testing.T) {
		dropRate := 0

		// FIXME: test times out w/ failure when providing "LAF"
		state0 := SetupSim("LAF", map[string]string{"--debuglog": ""}, 9, 0, 0, t)
		ticker := WatchMessageLists()

		if dropRate > 0 {
			state0.LogPrintf(logName, "DROP_RATE:%v", dropRate)
			RunCmd(fmt.Sprintf("S%v", dropRate))
		}

		stop := func() {
			ShutDownEverything(t)
			WaitForAllNodes(state0)
			ticker.Stop()
		}

		t.Run("Create Chain", func(t *testing.T) {
			e := factom.Entry{
				ChainID: id,
				ExtIDs:  extids,
				Content: encode("Hello World!"),
			}

			c := factom.NewChain(&e)

			commit, _ := ComposeChainCommit(a.Priv, c)
			reveal, _ := ComposeRevealEntryMsg(a.Priv, c.FirstEntry)

			state0.APIQueue().Enqueue(commit)
			state0.APIQueue().Enqueue(reveal)

			t.Run("Fund ChainCommit Address", func(t *testing.T) {
				amt := uint64(11)
				engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())
				WaitForAnyDeposit(state0, a.EcPub())
			})
		})

		t.Run("Generate Entries in Batches", func(t *testing.T) {
			WaitForZero(state0, a.EcPub())
			GenerateCommitsAndRevealsInBatches(t, state0)
		})

		t.Run("End simulation", func(t *testing.T) {
			WaitForZero(state0, a.EcPub())
			ht := state0.GetDBHeightComplete()
			WaitBlocks(state0, 2)
			newHt := state0.GetDBHeightComplete()
			assert.True(t, ht < newHt, "block height should progress")
			stop()
		})

	})
}

func GenerateCommitsAndRevealsInBatches(t *testing.T, state0 *state.State) {

	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	// KLUDGE vars duplicated from original test - should refactor
	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := GetBankAccount()

	batchCount := 1
	setDelay := 0     // blocks to wait between sets of entries
	numEntries := 250 // set the total number of entries to add

	logName := "simTest"
	state0.LogPrintf(logName, "BATCHES:%v", batchCount)
	state0.LogPrintf(logName, "ENTRIES:%v", numEntries)
	state0.LogPrintf(logName, "DELAY_BLOCKS:%v", setDelay)

	var batchTimes map[int]time.Duration = map[int]time.Duration{}

	for BatchID := 0; BatchID < int(batchCount); BatchID++ {

		publish := func(i int) {

			extids := [][]byte{encode(fmt.Sprintf("batch%v", BatchID))}

			e := factom.Entry{
				ChainID: id,
				ExtIDs:  extids,
				Content: encode(fmt.Sprintf("batch %v, seq: %v", BatchID, i)), // ensure no duplicate msg hashes
			}
			i++

			commit, _ := ComposeCommitEntryMsg(a.Priv, e)
			reveal, _ := ComposeRevealEntryMsg(a.Priv, &e)

			state0.APIQueue().Enqueue(commit)
			state0.APIQueue().Enqueue(reveal)
		}

		t.Run(fmt.Sprintf("Create Entries Batch %v", BatchID), func(t *testing.T) {

			tstart := WaitForEmptyHolding(state0, fmt.Sprintf("WAIT_HOLDING_START%v", BatchID))

			for x := 0; x < numEntries; x++ {
				publish(x)
			}

			t.Run("Fund EC Address", func(t *testing.T) {
				amt := uint64(numEntries)
				engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())
				//waitForAnyDeposit(state0, a.EcPub())
			})

			tend := WaitForEmptyHolding(state0, fmt.Sprintf("WAIT_HOLDING_END%v", BatchID))

			batchTimes[BatchID] = tend.Sub(tstart)

			state0.LogPrintf(logName, "BATCH %v RUNTIME %v", BatchID, batchTimes[BatchID])

			t.Run("Verify Entries", func(t *testing.T) {

				var sum time.Duration = 0

				for _, t := range batchTimes {
					sum = sum + t
				}

				if setDelay > 0 {
					WaitBlocks(state0, int(setDelay)) // wait between batches
				}

				//tend := waitForEmptyHolding(state0, fmt.Sprintf("SLEEP", BatchID))
				//bal := engine.GetBalanceEC(state0, a.EcPub())
				//assert.Equal(t, bal, int64(0))
				//assert.Equal(t, 0, len(state0.Holding), "messages stuck in holding")
			})
		})

	}

}
