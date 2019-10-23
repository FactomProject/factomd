package nettest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

type remoteNode struct {
	address string
	rpcPort int
}
type testNode struct {
	state       *state.State
	fnodes      map[int]remoteNode
	seedAddress string
}

var DEV_NET string = "127.0.0.1:18110"      // devnet port forward to fnode0 peer
var DOCKER_NETWORK string = "10.7.0.1:8110" // docker network
var SINGLE_NODE string = "127.0.0.1:39001"  // local netTest simulator

var DEV_NET_CONFIG string = `
LocalSpecialPeers                     = "127.0.0.1:18110"
Exclusive                             =  true 
`

func SetupNode(seedNode string, minPeers int, t *testing.T) *testNode {

	homeDir := testHelper.ResetSimHome(t)

	var CmdLineOptions map[string]string

	// Use identity 9 to run a follower
	if seedNode == DOCKER_NETWORK {
		testHelper.WriteConfigFile(9, 0, "", t)

		// use config that mirrors docker-compose from ./support/dev/docker-compose
		CmdLineOptions = map[string]string{
			"--db":                  "Map",
			"--network":             "CUSTOM",
			"--customnet":           "net",
			"--enablenet":           "true",
			"--blktime":             "30",
			"--count":               "1",
			"--startdelay":          "0",
			"--stdoutlog":           "out.txt",
			"--stderrlog":           "out.txt",
			"--checkheads":          "false",
			"--controlpanelsetting": "readwrite",
			"--debuglog":            "",
			"--logPort":             "39000",
			"--port":                "39001",
			"--controlpanelport":    "39002",
			"--networkport":         "39003",
			"--peers":               seedNode,
			"--factomhome":          homeDir,
		}
	}

	// when interfacing w/ devnet don't talk to other peers
	if seedNode == DEV_NET {
		testHelper.WriteConfigFile(9, 0, DEV_NET_CONFIG, t)

		CmdLineOptions = map[string]string{
			"--db":               "Map",
			"--network":          "LOCAL",
			"--enablenet":        "true",
			"--blktime":          "30",
			"--debuglog":         "",
			"--logPort":          "39000",
			"--port":             "39001",
			"--controlpanelport": "39002",
			"--networkport":      "39003",
			"--factomhome":       homeDir,
		}
	}

	n := &testNode{state: testHelper.StartPeer(CmdLineOptions)}

	t.Log("Waiting for first block")
	testHelper.WaitForBlock(n.state, 1)
	n.seedAddress = seedNode
	t.Log("Discovering Peers")
	n.DiscoverPeers(minPeers)
	return n
}

// status every minute for this local node
func (n *testNode) StatusEveryMinute() {
	// REVIEW: this could hit the RPC port to pull summary from remote nodes
	// instead of simply reporting on local fnode
	testHelper.StatusEveryMinute(n.state)
}

// wait minutes on this local node
func (n *testNode) WaitMinutes(minutes int) {
	testHelper.WaitMinutes(n.state, minutes)
}

// wait blocks on this local node
func (n *testNode) WaitBlocks(blocks int) {
	testHelper.WaitBlocks(n.state, blocks)
}

// wait for block on this local node
func (n *testNode) WaitForBlock(block int) {
	testHelper.WaitForBlock(n.state, block)
}

func (n *testNode) DiscoverPeers(minPeers int) {
	peers := n.GetPeers()

	for len(peers) < minPeers {
		peers = n.GetPeers()
		testHelper.WaitBlocks(n.state, 1) // let more time pass to discover peers
	}

	switch n.seedAddress {
	case SINGLE_NODE: // local simulator
		if minPeers != 0 {
			panic("local testing only")
		}
		n.fnodes = map[int]remoteNode{
			0: remoteNode{"127.0.0.1", 39001},
		}
	case DOCKER_NETWORK: // using ./support/dev/docker-compose.json
		n.fnodes = map[int]remoteNode{
			0: remoteNode{"10.7.0.1", 8088},
			1: remoteNode{"10.7.0.2", 8088},
			2: remoteNode{"10.7.0.3", 8088},
		}
	case DEV_NET: // using devnet port forwards
		if minPeers != 0 {
			panic("local testing only")
		}
		n.fnodes = map[int]remoteNode{
			0: remoteNode{"127.0.0.1", 8000},
			1: remoteNode{"127.0.0.1", 8001},
			2: remoteNode{"127.0.0.1", 8002},
			3: remoteNode{"127.0.0.1", 8003},
			4: remoteNode{"127.0.0.1", 8004},
		}
	default:
		/*
		   	REVIEW: automatic peer discovery should be re-visited after DevNet setup includes VPN
		       main blocker currently is that DevNet is setup to forward ports to loopback
		       so we cannot target the RPC api using this discovered IP Address
		*/
		panic("netTest only support hardcoded networks")
	}
}

// get list of peers from our local node
func (n *testNode) GetPeers() map[string]p2p.Peer {
	return n.state.NetworkController.GetKnownPeers()
}

// build url for debug API
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
	res, err := rpc(url, "network-info", "{}")

	if err != nil {
		panic(err)
	}

	json.Unmarshal(*res["result"], &info)

	return info
}

// FIXME: add a return code for waiting
func (r remoteNode) WaitForBlock(newBlock int) {
	rpc(r.getAPIUrl(), "wait-for-block", fmt.Sprintf(`{ "block": %v }`, newBlock))
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

func (r remoteNode) RunCmd(cmd string) error {
	_, err := rpc(r.getAPIUrl(), "sim-ctrl", fmt.Sprintf(`{ "Commands": ["%v"] }`, cmd))
	return err
}

func (r remoteNode) WriteConfig(identityNumber int, extra string) error {
	c := make(map[string]string)
	c["Config"] = testHelper.GetConfig(identityNumber, extra)
	cfg, _ := json.Marshal(c)
	_, err := rpc(r.getAPIUrl(), "write-configuration", string(cfg))
	return err
}

// make HTTP post
func postRequest(debugUrl string, jsonStr string) (*http.Response, error) {
	req, err := http.NewRequest("POST", debugUrl, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "text/plain;")
	req.Host = "localhost:8088" // KLUDGE: mux routing expects this to match the bound port

	client := &http.Client{}
	return client.Do(req)
}

// make calls to the debug api
func rpc(apiUrl string, method string, params string) (response map[string]*json.RawMessage, err error) {
	fmt.Printf("\nsend => %s", params)
	r, err := postRequest(apiUrl, fmt.Sprintf(`{"jsonrpc": "2.0", "id": 0, "method": "%s", "params": %s}`, method, params))
	defer r.Body.Close()

	if err != nil {
		return response, err
	}

	body, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("\nrpc => %s", body)

	err = json.Unmarshal(body, &response)
	return response, err
}
