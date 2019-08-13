package testHelper

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/wsapi"
)

const defaultTestPort = 8080

func InitTestState() {
	state := CreateAndPopulateTestStateAndStartValidator()
	state.SetPort(defaultTestPort)

	if wsapi.Servers == nil {
		wsapi.Servers = make(map[string]*wsapi.Server)
	}

	port := strconv.Itoa(defaultTestPort)
	if wsapi.Servers[port] == nil {
		server := wsapi.InitServer(state)
		wsapi.Servers[port] = server
	}
}

func getAPIUrl() string {
	return "http://localhost:" + fmt.Sprint(engine.GetFnodes()[0].State.GetPort()) + "/debug"
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

func SetInputFilter(apiRegex string) (*http.Response, error) {
	return postRequest(`{"jsonrpc": "2.0", "id": 0, "method": "message-filter", "params":{"output-regex":"", "input-regex":"` + apiRegex + `"}}`)
}

func SetOutputFilter(apiRegex string) (*http.Response, error) {
	return postRequest(`{"jsonrpc": "2.0", "id": 0, "method": "message-filter", "params":{"output-regex":"` + apiRegex + `", "input-regex":""}}`)
}

func DebugCall(method string, params string) (*http.Response, error){
	return postRequest(fmt.Sprintf(`{"jsonrpc": "2.0", "id": 0, "method": "%s", "params":%s}`, method, params))
}
