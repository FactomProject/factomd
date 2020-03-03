package simtest

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/stretchr/testify/assert"

	"github.com/FactomProject/factom"
)

// this applies chain & entry creation in 'proper' chronological order
func TestEntryBatch(t *testing.T) {

	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	extids := [][]byte{encode("foo"), encode("bar")}

	a := simulation.AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := simulation.AccountFromFctSecret("Fs2BNvoDgSoGJpWg4PvRUxqvLE28CQexp5FZM9X5qU6QvzFBUn6D")

	numEntries := 9 // set the total number of entries to add

	println(a.String())

	state0 := simulation.SetupSim("LLAAFF", nil, 10, 0, 0, t)

	var entries []interfaces.IMsg
	var oneFct uint64 = factom.FactoidToFactoshi("1")
	var ecMargin = 100 // amount of ec to have left

	{ // fund entries & chain create
		simulation.WaitForZeroEC(state0, a.EcPub()) // assert we are starting from zero

		b.FundFCT(oneFct * 20)                          // transfer coinbase funds to b
		b.SendFCT(a, oneFct*10)                         // use account b to fund a.ConvertEC() from above
		a.ConvertEC(uint64(numEntries + 11 + ecMargin)) // Chain costs 10 + 1 per k so our chain head costs 11

		simulation.WaitForEcBalanceOver(state0, a.EcPub(), int64(ecMargin-1)) // wait for all entries to process
	}

	{ // create the chain
		e := factom.Entry{
			ChainID: id,
			ExtIDs:  extids,
			Content: encode("Hello World!"),
		}

		c := factom.NewChain(&e)

		commit, _ := simulation.ComposeChainCommit(a.Priv, c)
		reveal, _ := simulation.ComposeRevealEntryMsg(a.Priv, c.FirstEntry)

		state0.APIQueue().Enqueue(commit)
		state0.APIQueue().Enqueue(reveal)
	}

	simulation.WaitMinutes(state0, 1)

	{ // write entries

		for i := 0; i < numEntries; i++ {
			e := factom.Entry{
				ChainID: id,
				ExtIDs:  extids,
				Content: encode(fmt.Sprintf("hello@%v", i)), // ensure no duplicate msg hashes
			}
			commit, _ := simulation.ComposeCommitEntryMsg(a.Priv, e)
			reveal, _ := simulation.ComposeRevealEntryMsg(a.Priv, &e)

			state0.LogMessage("simtest", "commit", commit)
			state0.LogMessage("simtest", "reveal", reveal)

			entries = append(entries, commit)
			entries = append(entries, reveal)

			state0.APIQueue().Enqueue(commit)
			state0.APIQueue().Enqueue(reveal)
		}

	}

	simulation.WaitForEcBalanceUnder(state0, a.EcPub(), int64(ecMargin+1)) // wait for all entries to process
	simulation.WaitBlocks(state0, 1)                                       // give time for holding to clear

	simulation.ShutDownEverything(t)
	simulation.WaitForAllNodes(state0)

	assert.Equal(t, int64(ecMargin), a.GetECBalance()) // should have 100 extra EC's

	/*
		{ // check outputs
			assert.Equal(t, int64(0), a.GetECBalance())

			for _, fnode := range fnode.GetFnodes() {
				s := fnode.State
				for _, h := range state0.Hold.Messages() {
					for _, m := range h {
						s.LogMessage("dependentHolding", "stuck", m)
					}
				}
				assert.Equal(t, 0, len(s.Holding), "messages stuck in holding")
				assert.Equal(t, 0, s.Hold.GetSize(), "messages stuck in New Holding")
			}

		}
	*/
}
