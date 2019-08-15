package nettest

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

type remoteNode struct {
	address string
}
type testNode struct {
	state  *state.State
	fnodes map[int]remoteNode
}

// extract last number of ipv4 as an int
func ipLastOctet(addr string) int {
	i, _ := strconv.ParseInt(strings.Split(addr, ".")[3], 10, 64)
	return int(i)
}

func SetupNode(seedNode string, t *testing.T) testNode {

	homeDir := testHelper.ResetSimHome(t)

	// Use identity 9
	testHelper.WriteConfigFile(9, 0, "", t)

	// use config that mirrors docker-compose from ./support/dev/docker-compose
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
		"--peers":               seedNode,
		"--factomhome":          homeDir,
	}

	n := testNode{state: testHelper.StartSim(1, CmdLineOptions)}

	testHelper.WaitForBlock(n.state, 1)

	peers := n.GetPeers()

	for len(peers) == 0 {
		peers = n.GetPeers()
		testHelper.WaitBlocks(n.state, 1) // let more time pass to discover peers
	}

	n.fnodes = make(map[int]remoteNode)

	for _, p := range peers {
		// REVIEW: may need to set node order by port or
		// maybe a config setting from the target node instead
		i := ipLastOctet(p.Address)

		// offset from 0 - ip's usually won't start at 0
		n.fnodes[i-1] = remoteNode{address: p.Address}
	}

	return n
}

func postRequest(debugUrl string, jsonStr string) (*http.Response, error) {
	req, err := http.NewRequest("POST", debugUrl, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "text/plain;")

	client := &http.Client{}
	return client.Do(req)
}

// make calls to the debug api
func rpc(apiUrl string, method string, params string) []byte {
	r, err := postRequest(apiUrl, fmt.Sprintf(`{"jsonrpc": "2.0", "id": 0, "method": "%s", "params":%s}`, method, params))
	defer r.Body.Close()

	if err != nil { panic(err) }

	body, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("BODY: %s", body)

	// FIXME add better error handling
	// for example wait-for-block in past triggers an error & returns w/ empty body
	return body
}

func (n testNode) GetPeers() map[string]p2p.Peer {
	return n.state.NetworkController.GetKnownPeers()
}

func (r remoteNode) getAPIUrl() string {
	if r.address == "" {
		panic("remoteNode is nil")
	}
	return "http://" + r.address + ":8088/debug"
}

func (r remoteNode) WaitForBlock(newBlock int) {
	url := r.getAPIUrl()
	rpc(url,"wait-for-block", fmt.Sprintf(`{ "block": %v }`, newBlock))
}

func (r remoteNode) WaitBlocks(blks int) {
	rpc(r.getAPIUrl(), "wait-blocks", fmt.Sprintf(`{ "blocks": %v }`, blks))
}

func (r remoteNode) WaitForMinute(minute int) {
	rpc(r.getAPIUrl(), "wait-for-minute", fmt.Sprintf(`{ "minute": %v }`, minute))
}

func (r remoteNode) WaitMinutes(minutes int) {
	rpc(r.getAPIUrl(), "wait-minutes", fmt.Sprintf(`{ "minutes": %v }`, minutes))
}
