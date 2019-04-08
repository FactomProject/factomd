package simtest

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/engine"
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
	a := AccountFromFctSecret("Fs2zQ3egq2j99j37aYzaCddPq9AF3mgh64uG9gRaDAnrkjRx3eHs")
	b := GetBankAccount()

	numEntries := 250 // set the total number of entries to add

	println(b.String()) // print factoid & ec addresses
	println(a.String())

	params := map[string]string{"--debuglog": ""}
	state0 := SetupSim("L", params, 8, 0, 0, t)

	publish := func(i int) {
		e := factom.Entry{
			ChainID: id,
			ExtIDs:  extids,
			Content: encode(fmt.Sprintf("hello@%v", i)), // ensure no duplicate msg hashes
		}
		commit, _ := ComposeCommitEntryMsg(a.Priv, e)
		reveal, _ := ComposeRevealEntryMsg(a.Priv, &e)

		state0.APIQueue().Enqueue(commit)
		state0.APIQueue().Enqueue(reveal)
	}

	for x := 0; x < numEntries; x++ {
		publish(x)
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

	amt := uint64(numEntries + 11) // Chain costs 10 + 1 per k so our chain costs 11
	engine.FundECWallet(state0, b.FctPrivHash(), a.EcAddr(), amt*state0.GetFactoshisPerEC())
	WaitForAnyDeposit(state0, a.EcPub())

	WaitForZero(state0, a.EcPub())
	ShutDownEverything(t)
	WaitForAllNodes(state0)

	bal := engine.GetBalanceEC(state0, a.EcPub())
	//fmt.Printf("Bal: => %v", bal)
	assert.Equal(t, int64(0), bal)

	for _, v := range state0.Holding {
		s, _ := v.JSONString()
		println(s)
	}

	assert.Equal(t, 0, len(state0.Holding), "messages stuck in holding")

}
