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
	rpcPort	int
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
		//fmt.Printf("%v", peers)
		testHelper.WaitBlocks(n.state, 1) // let more time pass to discover peers
	}

	/*
	REVIEW: automatic peer discovery should be re-visited after DevNet setup includes VPN

	n.fnodes = make(map[int]remoteNode)
	n.fnodes[0] = remoteNode{seedNode}

	for _, p := range peers {
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
	 */

	switch seedNode {
	case "10.7.0.1:8110": // docker network
		t.Logf("Using Hardcoded Docker peers")

		n.fnodes = map[int]remoteNode{
			0: remoteNode{"10.7.0.1", 8088},
			1: remoteNode{"10.7.0.2", 8088},
			2: remoteNode{"10.7.0.3", 8088},
		}
	case "127.0.0.1:39001": // local simulator
		if minPeers != 0 {
			panic("local testing only")
		}
		n.fnodes = map[int]remoteNode{
			0: remoteNode{"127.0.0.1", 39001},
		}
	default:
		panic("netTest only support hardcoded networks")
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
func rpc(apiUrl string, method string, params string) (response map[string]*json.RawMessage, err error) {
	r, err := postRequest(apiUrl, fmt.Sprintf(`{"jsonrpc": "2.0", "id": 0, "method": "%s", "params":%s}`, method, params))

	if err != nil {
		return response, err
	}

	defer r.Body.Close()

	body, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("\nrpc => %s", body)

	// FIXME add better error handling
	// for example wait-for-block in past triggers an error & returns w/ empty body
	err = json.Unmarshal(body, &response)
	return response, err
}

func (n testNode) GetPeers() map[string]p2p.Peer {
	return n.state.NetworkController.GetKnownPeers()
}

func (r remoteNode) getAPIUrl() string {
	if r.address == "" {
		panic("remoteNode is nil")
	}
	return fmt.Sprintf("http://%s:%v/debug", r.address, r.rpcPort)
}

type netInfo struct {
	NodeName      string
	Role          string
	NetworkNumber int
	NetworkName   string
	NetworkID     uint32
}

func (r remoteNode) NetworkInfo() (info netInfo) {
	url := r.getAPIUrl()
	res, err := rpc(url,"network-info","{}")

	if err != nil {
		panic(err)
	}

	json.Unmarshal(*res["result"], &info)

	return info
}

// FIXME: add a return code for waiting
func (r remoteNode) WaitForBlock(newBlock int) {
	rpc(r.getAPIUrl(),"wait-for-block", fmt.Sprintf(`{ "block": %v }`, newBlock))
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

func (r remoteNode) RunCmd(cmd string) {
	rpc(r.getAPIUrl(), "sim-ctrl", fmt.Sprintf(`{ "Commands": ["%v"] }`, cmd))
}
