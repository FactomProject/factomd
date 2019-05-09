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

func TestHoldingRebound(t *testing.T) {
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

	println(b.String())
	println(a.String())

	dropRate := 0

	params := map[string]string{"--debuglog": ""}

	// REVIEW: changing simulation to LAF doesn't always pass cleanly on circle
	// this may just mean we shouldn't check for empty holding
	state0 := SetupSim("L", params, 9, 0, 0, t)
	ticker := WatchMessageLists()
	defer ticker.Stop()

	if dropRate > 0 {
		state0.LogPrintf(logName, "DROP_RATE:%v", dropRate)
		RunCmd(fmt.Sprintf("S%v", dropRate))
	}

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

	amt := uint64(11)
	engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())
	WaitForAnyDeposit(state0, a.EcPub())

	WaitForZero(state0, a.EcPub())
	GenerateCommitsAndRevealsInBatches(t, state0)
	WaitForZero(state0, a.EcPub())
	ht := state0.GetDBHeightComplete()
	WaitBlocks(state0, 2)
	newHt := state0.GetDBHeightComplete()
	assert.True(t, ht < newHt, "block height should progress")

	ShutDownEverything(t)
	WaitForAllNodes(state0)
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

		tstart := WaitForEmptyHolding(state0, fmt.Sprintf("WAIT_HOLDING_START%v", BatchID))

		for x := 0; x < numEntries; x++ {
			publish(x)
		}

		amt := uint64(numEntries)
		engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())
		WaitForAnyDeposit(state0, a.EcPub())

		tend := WaitForEmptyHolding(state0, fmt.Sprintf("WAIT_HOLDING_END%v", BatchID))

		batchTimes[BatchID] = tend.Sub(tstart)

		state0.LogPrintf(logName, "BATCH %v RUNTIME %v", BatchID, batchTimes[BatchID])

		var sum time.Duration = 0

		for _, t := range batchTimes {
			sum = sum + t
		}

		if setDelay > 0 {
			WaitBlocks(state0, int(setDelay)) // wait between batches
		}

		// FIXME: do a real test
		// this may mean allowing for some types of entries to remain in holding

		//tend := waitForEmptyHolding(state0, fmt.Sprintf("SLEEP", BatchID))
		//bal := engine.GetBalanceEC(state0, a.EcPub())
		//assert.Equal(t, bal, int64(0))
		//assert.Equal(t, 0, len(state0.Holding), "messages stuck in holding")
	}
}
