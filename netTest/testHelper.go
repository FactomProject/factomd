package nettest

import (
	"bytes"
	"encoding/json"
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

func SetupNode(seedNode string, minPeers int, t *testing.T) testNode {

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

	for len(peers) < minPeers {
		peers = n.GetPeers()
		fmt.Printf("%v", peers)
		testHelper.WaitBlocks(n.state, 1) // let more time pass to discover peers
	}

	// KLUDGE: always keep the seed peer
	n.fnodes = make(map[int]remoteNode)
	n.fnodes[0] = remoteNode{seedNode}

	for _, p := range peers {
		// REVIEW: may need to set node order by port or
		// maybe a config setting from the target node instead
		i := ipLastOctet(p.Address)

		// offset from 0 - ip's usually won't start at 0
		r := remoteNode{address: p.Address}
		info, err := r.NetworkInfo()
		if err == nil {
			// keep this node
			_ = info
			n.fnodes[i-1] = r
		}
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
func rpc(apiUrl string, method string, params string) (body []byte, err error) {
	r, err := postRequest(apiUrl, fmt.Sprintf(`{"jsonrpc": "2.0", "id": 0, "method": "%s", "params":%s}`, method, params))

	if err != nil {
		return body, err
	}

	defer r.Body.Close()

	body, _ = ioutil.ReadAll(r.Body)
	fmt.Printf("\nrpc => %s", body)

	// FIXME add better error handling
	// for example wait-for-block in past triggers an error & returns w/ empty body
	return body, err
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

func (r remoteNode) NetworkInfo() (info map[string]interface{}, err error) {
	url := r.getAPIUrl()
	res, err := rpc(url,"network-info","{}")

	if err !=nil {
		return info, err
	}

	info = make(map[string]interface{})
	json.Unmarshal(res, info)
	fmt.Sprintf("\nINFO: %v", info)

	return info, nil
}

// FIXME: add a return code for waiting
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
