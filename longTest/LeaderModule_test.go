package longtest

import (
	"bytes"
	"testing"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/fnode"
	. "github.com/FactomProject/factomd/testHelper"
)

//  Currently just tests FCT/EC Commit/Reveal messages to make sure leader is working
func TestLeaderModule(t *testing.T) {

	encode := func(s string) []byte {
		b := bytes.Buffer{}
		b.WriteString(s)
		return b.Bytes()
	}

	id := "92475004e70f41b94750f4a77bf7b430551113b25d3d57169eadca5692bb043d"
	extids := [][]byte{encode("foo"), encode("bar")}

	//a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := AccountFromFctSecret("Fs2BNvoDgSoGJpWg4PvRUxqvLE28CQexp5FZM9X5qU6QvzFBUn6D")

	numEntries := 1 // set the total number of entries to add

	//println(a.String())

	state0 := SetupSim("LF", nil, 10, 0, 0, t)

	//var entries []interfaces.IMsg
	var oneFct uint64 = factom.FactoidToFactoshi("1")
	var ecMargin = 100 // amount of ec to have left

	WaitBlocks(state0, 1)
	state1 := fnode.Get(1).State

	{ // fund entries & chain create
		//WaitForZeroEC(state0, a.EcPub()) // assert we are starting from zero

		b.FundFCT(oneFct * 20) // transfer coinbase funds to b
		WaitBlocks(state0, 1)
		b.ConvertEC(uint64(numEntries + 11 + ecMargin)) // Chain costs 10 + 1 per k so our chain head costs 11
		WaitForEcBalanceOver(state0, b.EcPub(), 1)      // wait for all entries to process
	}

	{ // create the chain
		e := factom.Entry{
			ChainID: id,
			ExtIDs:  extids,
			Content: encode("Hello World!"),
		}

		c := factom.NewChain(&e)

		commit, _ := ComposeChainCommit(b.Priv, c)
		reveal, _ := ComposeRevealEntryMsg(b.Priv, c.FirstEntry)

		state0.APIQueue().Enqueue(commit)
		WaitBlocks(state0, 1) // make sure commit is in holding when reveal goes through
		// since old leader behavior would execute commit/reveal matches
		state0.APIQueue().Enqueue(reveal)

		state0.LogMessage("simtest", "pushing into API", commit)
		state0.LogPrintf("simtest", "pushing into API hash : %v", commit.GetHash())
		state0.LogPrintf("simtest", "pushing into API hash : %v", reveal.GetHash())
		WaitForEntry(state0, commit.GetHash())
	}

	{ // KLUDGE debug
		WaitForBlock(state1, 4)
		ShutDownEverything(t)
	}

}
