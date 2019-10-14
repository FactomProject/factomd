package simtest

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestHoldingRebound(t *testing.T) {
	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	extids := [][]byte{encode("foo"), encode("bar")}
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")

	println(a.String())
	state0 := SetupSim("L", map[string]string{}, 120, 0, 0, t)

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

	a.FundEC(11)
	GenerateCommitsAndRevealsInBatches(t, state0)

	ht := state0.GetDBHeightComplete()
	WaitBlocks(state0, 2)
	newHt := state0.GetDBHeightComplete()
	assert.True(t, ht < newHt, "block height should progress")

	ShutDownEverything(t)
	WaitForAllNodes(state0)

	for _, ml := range state0.Hold.Messages() {
		for _, m := range ml {
			state0.LogMessage("simTest", "stuck", m)
		}
	}
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

	batchCount := 10
	setDelay := 1     // blocks to wait between sets of entries
	numEntries := 250 // set the total number of entries to add

	logName := "simTest"
	state0.LogPrintf(logName, "BATCHES:%v", batchCount)
	state0.LogPrintf(logName, "ENTRIES:%v", numEntries)
	state0.LogPrintf(logName, "DELAY_BLOCKS:%v", setDelay)

	var batchTimes = make(map[int]time.Duration)

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

		for x := 0; x < numEntries; x++ {
			publish(x)
		}

		time.Sleep(10*time.Second)
		{ // measure time it takes to process all messages by observing entry credit spend
			tstart := time.Now()
			a.FundEC(uint64(numEntries + 1))
			WaitForEcBalanceUnder(state0, a.EcPub(), int64(BatchID+2))
			tend := time.Now()
			batchTimes[BatchID] = tend.Sub(tstart)
			state0.LogPrintf(logName, "BATCH %v RUNTIME %v", BatchID, batchTimes[BatchID])
		}

		if setDelay > 0 {
			WaitBlocks(state0, int(setDelay)) // wait between batches
		}
	}
}
