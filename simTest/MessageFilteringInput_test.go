package simtest

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestFilterAPIInput(t *testing.T) {

	state0 := SetupSim("LLLLLAAF", map[string]string{"--debuglog": "."}, 25, 1, 1, t)

	RunCmd("1")
	RunCmd("w")
	RunCmd("s")

	apiRegex := "EOM.*5/.*minute 1"

	// API call
	url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
	jsonStr := []byte(`{"jsonrpc": "2.0", "id": 0, "method": "message-filter", "params":{"output-regex":"", "input-regex":"` + apiRegex + `"}}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("content-type", "text/plain;")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	WaitBlocks(state0, 5)

	// The message-filter call we did above should have caused an election and SO, Node01 should not be a leader anymore.
	if GetFnodes()[1].State.Leader {
		t.Fatalf("Node01 should not be leader!")
	}

	CheckAuthoritySet(t)

	// Check Node01 Network Input logs to make sure there are no enqueued including our Regex
	out := SystemCall(`grep "enqueue" fnode01_networkinputs.txt | grep "` + apiRegex + `" | grep -v "EmbeddedMsg" | wc -l`)

	if strings.TrimSuffix(strings.Trim(string(out), " "), "\n") != string("0") {
		t.Fatalf("Filter missed let a message pass 1.")
	}

	// Check Node01 Network Input logs to make sure there are no Dropped messaged besides the ones for our Regex
	out2 := SystemCall(`grep "Drop, matched filter Regex" fnode01_networkinputs.txt | grep -v "` + apiRegex + `" | wc -l`)

	if strings.TrimSuffix(strings.Trim(string(out2), " "), "\n") != string("0") {
		t.Fatalf("Filter missed let a message pass 2.")
	}

	WaitBlocks(state0, 1)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}
}
