package simulation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/wsapi"
)

/*
Don't expose wsapi references directly to simulator
*/

type EntryRequest = wsapi.EntryRequest
type MessageRequest = wsapi.MessageRequest
type GeneralTransactionData = wsapi.GeneralTransactionData
type EntryStatus = wsapi.EntryStatus
type TransactionRequest = wsapi.TransactionRequest

var HandleV2RawData = wsapi.HandleV2RawData
var HandleV2SendRawMessage = wsapi.HandleV2SendRawMessage
var SetState = wsapi.SetState

func V2Request(req *primitives.JSON2Request, port int) (*primitives.JSON2Response, error) {
	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	portStr := fmt.Sprintf("%d", port)
	resp, err := http.Post(
		"http://localhost:"+portStr+"/v2",
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := primitives.NewJSON2Response()
	if err := json.Unmarshal(body, r); err != nil {
		return nil, err
	}
	return nil, nil
}
