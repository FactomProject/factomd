package nettest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

type testNode struct {
	state *state.State
}

var default_ip string = "10.7.0.1" // KLUDGE: this is the network address from docker

func SetupNode(t *testing.T) testNode {

	homeDir := testHelper.ResetSimHome(t)

	// Use identity 9
	testHelper.WriteConfigFile(9, 0, "", t)

	CmdLineOptions := map[string]string{
		"--db":                  "Map",
		"--network":             "CUSTOM",
		"--customnet":           "net",
		"--enablenet":           "true",
		"--blktime":             "15",
		"--count":               "1",
		"--startdelay":          "0",
		"--stdoutlog":           "out.txt",
		"--stderrlog":           "out.txt",
		"--checkheads":          "false",
		"--controlpanelsetting": "readwrite",
		"--debuglog":            "faulting|bad",
		"--logPort":             "39000",
		"--port":                "39001",
		"--controlpanelport":    "39002",
		"--networkport":         "39003",
		"--peers":               fmt.Sprintf("%s:8110", default_ip),
		"--factomhome":          homeDir,
	}

	n := testNode{testHelper.StartSim(1, CmdLineOptions)}

	testHelper.WaitForBlock(n.state, 1) // wait until we are processing blocks

	return n
}

func getAPIUrl() string {
	return "http://" + default_ip + ":8088/debug"
}

func postRequest(jsonStr string) (*http.Response, error) {
	req, err := http.NewRequest("POST", getAPIUrl(), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "text/plain;")

	client := &http.Client{}
	return client.Do(req)
}

// make calls to the debug api
func rpc(method string, params string) []byte {
	r, _ := postRequest(fmt.Sprintf(`{"jsonrpc": "2.0", "id": 0, "method": "%s", "params":%s}`, method, params))
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("BODY: %s", body)

	// FIXME  add better error handling
	// for example wait-for-block in past triggers an error & returns w/ empty body
	return body
}

// FIXME: these functions need to be able to specify which node is being waited on

func (n testNode) GetPeers() map[string]p2p.Peer {
	return n.state.NetworkController.GetKnownPeers()
}

func (testNode) WaitForBlock(newBlock int) {
	rpc("wait-for-block", fmt.Sprintf(`{ "block": %v }`, newBlock))
}

func (testNode) WaitBlocks(blks int) {
	rpc("wait-blocks", fmt.Sprintf(`{ "blocks": %v }`, blks))
}

func (testNode) WaitForMinute(minute int) {
	rpc("wait-for-minute", fmt.Sprintf(`{ "minute": %v }`, minute))
}

func (testNode) WaitMinutes(minutes int) {
	rpc("wait-minutes", fmt.Sprintf(`{ "minutes": %v }`, minutes))
}
