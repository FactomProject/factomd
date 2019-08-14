package nettest

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"io/ioutil"
	"net/http"
	"testing"
)

type testNode struct {
	*state.State
}

var default_ip string = "10.7.0.1" // KLUDGE: this is the network address from docker

func SetupNode(t *testing.T) testNode {

	homeDir := testHelper.ResetSimHome(t)

	// Use identity 9
	testHelper.WriteConfigFile(9, 0, "", t)

	CmdLineOptions := map[string]string{
		"--db":                  "Map",
		"--network":             "CUSTOM",
		"--customnet":            "net",
		"--enablenet":           "true",
		"--blktime":             "15",
		"--count":               "1",
		"--startdelay":          "1",
		"--stdoutlog":           "out.txt",
		"--stderrlog":           "out.txt",
		"--checkheads":          "false",
		"--controlpanelsetting": "readwrite",
		"--debuglog":            "faulting|bad",
		"--logPort":             "39000",
		"--port":                "39001",
		"--controlpanelport":    "39002",
		"--networkport":         "39003",
		"--peers":        fmt.Sprintf("%s:8110", default_ip),
		"--factomhome": homeDir,
	}

	n := testNode{testHelper.StartSim(1, CmdLineOptions)}

	testHelper.WaitForBlock(n.State, 1) // wait until we are processing blocks

	// TODO: get list of discovered peers we can hit & build a mapping of Fnodes

	return n
}

func getAPIUrl() string {
	return "http://"+default_ip+":8088/debug"
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
func rpc (method string, params string) {
	r, _ := postRequest(fmt.Sprintf(`{"jsonrpc": "2.0", "id": 0, "method": "%s", "params":%s}`, method, params))
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("BODY: %s", body)
}

