package simtest

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/engine"
	"testing"

	"github.com/FactomProject/factom"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
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

	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := AccountFromFctSecret("Fs2BNvoDgSoGJpWg4PvRUxqvLE28CQexp5FZM9X5qU6QvzFBUn6D")

	numEntries := 9 // set the total number of entries to add

	println(a.String())

	params := map[string]string{"--debuglog": ""}
	state0 := SetupSim("LLAAFF", params, 10, 0, 0, t)

	var entries []interfaces.IMsg

	{ // funding the chain + entries
		var oneFct uint64 = factom.FactoidToFactoshi("1")
		b.FundFCT(oneFct*20) // transfer coinbase funds to b
		b.SendFCT(a, oneFct*10) // use account b to fund a.ConvertEC() from above
		a.ConvertEC(uint64(numEntries + 11)) // buy EC's needed for all actions
		WaitForAnyEcBalance(state0, a.EcPub()) // wait for funding
	}

	{ // create the chain
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
	}

	{ // write entries

		for i := 0; i < numEntries; i++ {
			e := factom.Entry{
				ChainID: id,
				ExtIDs:  extids,
				Content: encode(fmt.Sprintf("hello@%v", i)), // ensure no duplicate msg hashes
			}
			commit, _ := ComposeCommitEntryMsg(a.Priv, e)
			reveal, _ := ComposeRevealEntryMsg(a.Priv, &e)

			state0.LogMessage("simtest", "commit", commit)
			state0.LogMessage("simtest", "reveal", reveal)

			entries = append(entries, commit)
			entries = append(entries, reveal)

			state0.APIQueue().Enqueue(commit)
			state0.APIQueue().Enqueue(reveal)
		}

		//WaitMinutes(state0, 2) // ensure messages are reviewed in holding at least once
		WaitForZeroEC(state0, a.EcPub()) // wait for all entries to be paid/written
	}

	WaitBlocks(state0, 1) // give time for holding to clear

	ShutDownEverything(t)
	WaitForAllNodes(state0)

	{ // check outputs
		assert.Equal(t, int64(0), a.GetECBalance())

		for _, fnode := range engine.GetFnodes() {
			s := fnode.State
			for _, h := range state0.Hold.Messages() {
				for _, m := range h {
					s.LogMessage("newholding", "stuck", m)
				}
			}
			assert.Equal(t, 0, len(s.Holding), "messages stuck in holding")
			assert.Equal(t, 0, s.Hold.GetSize(), "messages stuck in New Holding")
		}

	}
}
