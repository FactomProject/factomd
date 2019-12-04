package simtest

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/fnode"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factom"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
)

func TestEntriesBeforeChain(t *testing.T) {

	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	extids := [][]byte{encode("foo"), encode("bar")}
	var lastentry interfaces.IHash

	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := AccountFromFctSecret("Fs2BNvoDgSoGJpWg4PvRUxqvLE28CQexp5FZM9X5qU6QvzFBUn6D")

	numEntries := 9 // set the total number of entries to add

	state0 := SetupSim("LF", map[string]string{"--debuglog": "."}, 10, 0, 0, t)

	var entries []interfaces.IMsg
	var oneFct uint64 = factom.FactoidToFactoshi("1")
	var ecMargin = 100

	{ // publish entries
		publish := func(i int) {
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

		for x := 0; x < numEntries; x++ {
			publish(x)
		}

	}

	{ // create chain
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
		lastentry = reveal.Entry.GetHash()
	}

	// REVIEW is this a good enough test for holding
	WaitMinutes(state0, 2) // ensure messages are reviewed in holding at least once

	{ // fund FCT address & chain & entries

		WaitForZeroEC(state0, a.EcPub())
		// initially unfunded EC conversion
		a.ConvertEC(uint64(numEntries + 11 + ecMargin)) // Chain costs 10 + 1 per k so our chain head costs 11

		b.FundFCT(oneFct * 20)  // transfer coinbase funds to b
		b.SendFCT(a, oneFct*10) // use account b to fund a.ConvertEC() from above

		WaitForEcBalanceOver(state0, a.EcPub(), int64(ecMargin-1))
	}

	WaitBlocks(state0, 2) // give time for holding to clear
	WaitForEcBalanceUnder(state0, a.EcPub(), int64(ecMargin+1))
	WaitForEntry(state0, lastentry)

	Halt(t)
	//ShutDownEverything(t)
	//WaitForAllNodes(state0)

	assert.Equal(t, int64(ecMargin), a.GetECBalance())              // should have 100 extra EC's
	assert.Equal(t, a.GetECBalance(fnode.Get(1)), a.GetECBalance()) // both nodes should report the same values
}
