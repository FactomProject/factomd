package wsapi

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func (server *Server) AddV1Endpoints() {
	server.addRoute("/v1/factoid-submit/", HandleFactoidSubmit)
	server.addRoute("/v1/commit-chain/", HandleCommitChain)
	server.addRoute("/v1/reveal-chain/", HandleRevealChain)
	server.addRoute("/v1/commit-entry/", HandleCommitEntry)
	server.addRoute("/v1/reveal-entry/", HandleRevealEntry)

	server.addRoute("/v1/directory-block-head/", HandleDirectoryBlockHead)
	server.addRoute("/v1/get-raw-data/{hash}/", HandleGetRaw)
	server.addRoute("/v1/get-receipt/{hash}/", HandleGetReceipt)
	server.addRoute("/v1/directory-block-by-keymr/{keymr}/", HandleDirectoryBlock)
	server.addRoute("/v1/directory-block-height/", HandleDirectoryBlockHeight)
	server.addRoute("/v1/entry-block-by-keymr/{keymr}/", HandleEntryBlock)
	server.addRoute("/v1/entry-by-hash/{hash}/", HandleEntry)
	server.addRoute("/v1/chain-head/{chainid}/", HandleChainHead)
	server.addRoute("/v1/entry-credit-balance/{address}/", HandleEntryCreditBalance)
	server.addRoute("/v1/factoid-balance/{address}/", HandleFactoidBalance)
	server.addRoute("/v1/factoid-get-fee/", HandleGetFee)
	server.addRoute("/v1/properties/", HandleProperties)
	server.addRoute("/v1/heights/", HandleHeights)

	server.addRoute("/v1/dblock-by-height/{height:[0-9]+}/", HandleDBlockByHeight)
	server.addRoute("/v1/ecblock-by-height/{height:[0-9]+}/", HandleECBlockByHeight)
	server.addRoute("/v1/fblock-by-height/{height:[0-9]+}/", HandleFBlockByHeight)
	server.addRoute("/v1/ablock-by-height/{height:[0-9]+}/", HandleABlockByHeight)
	server.addRoute("/v1/dblock-by-height/{height:[0-9]+}/", HandleDBlockByHeight)
}

func extractURLHeightParam(writer http.ResponseWriter, request *http.Request) (param HeightRequest, err error) {
	params := mux.Vars(request)
	height := params["height"]

	h, err := strconv.ParseInt(height, 0, 64)
	if err != nil {
		return param, err
	}
	param = HeightRequest{Height: h}
	return param, err
}

func extractURLHashParam(request *http.Request) (param HashRequest) {
	params := mux.Vars(request)
	hashkey := params["hash"]
	return HashRequest{Hash: hashkey}
}

func extractURLKeyMRParam(request *http.Request) (param KeyMRRequest) {
	params := mux.Vars(request)
	keyMR := params["keymr"]
	return KeyMRRequest{KeyMR: keyMR}
}

func extractURLChainIDParam(request *http.Request) (param ChainIDRequest) {
	params := mux.Vars(request)
	chainID := params["chainid"]
	return ChainIDRequest{ChainID: chainID}
}

func extractURLAddressParam(request *http.Request) (param AddressRequest) {
	params := mux.Vars(request)
	address := params["address"]
	return AddressRequest{Address: address}
}

func HandleDBlockByHeight(writer http.ResponseWriter, request *http.Request) {
	param, err := extractURLHeightParam(writer, request)
	if err != nil {
		handleV1Error(writer, NewInvalidParamsError())
		return
	}
	req := primitives.NewJSON2Request("dblock-by-height", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	returnV1(writer, jsonResp, jsonError)
}

func HandleECBlockByHeight(writer http.ResponseWriter, request *http.Request) {
	param, err := extractURLHeightParam(writer, request)
	if err != nil {
		handleV1Error(writer, NewInvalidParamsError())
		return
	}
	req := primitives.NewJSON2Request("ecblock-by-height", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	returnV1(writer, jsonResp, jsonError)
}

func HandleFBlockByHeight(writer http.ResponseWriter, request *http.Request) {
	param, err := extractURLHeightParam(writer, request)
	if err != nil {
		handleV1Error(writer, NewInvalidParamsError())
		return
	}

	req := primitives.NewJSON2Request("fblock-by-height", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	returnV1(writer, jsonResp, jsonError)
}

func HandleABlockByHeight(writer http.ResponseWriter, request *http.Request) {
	param, err := extractURLHeightParam(writer, request)
	if err != nil {
		handleV1Error(writer, NewInvalidParamsError())
		return
	}

	req := primitives.NewJSON2Request("ablock-by-height", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	returnV1(writer, jsonResp, jsonError)
}

func HandleCommitChain(writer http.ResponseWriter, request *http.Request) {
	type commitchain struct {
		CommitChainMsg string
	}
	c := new(commitchain)
	if p, err := ioutil.ReadAll(request.Body); err != nil {
		handleV1Error(writer, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, c); err != nil {
			handleV1Error(writer, NewInvalidParamsError())
			return
		}
	}
	param := MessageRequest{Message: c.CommitChainMsg}
	req := primitives.NewJSON2Request("commit-chain", 1, param)
	_, jsonError := HandleV2Request(writer, request, req)

	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	//log.Print(jsonResp.Result.(*RevealEntryResponse).Message)

	// this is the blank '200 ok' that is returned for V1
	returnV1Msg(writer, "", false)

}

func HandleRevealChain(writer http.ResponseWriter, request *http.Request) {
	HandleRevealEntry(writer, request)
}

func HandleCommitEntry(writer http.ResponseWriter, request *http.Request) {
	type commitentry struct {
		CommitEntryMsg string
	}

	c := new(commitentry)
	if p, err := ioutil.ReadAll(request.Body); err != nil {
		handleV1Error(writer, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, c); err != nil {
			handleV1Error(writer, NewInvalidParamsError())
			return
		}
	}

	//  v2 wants a MessageRequest instead of an EntryRequest
	//	param := EntryRequest{Entry: c.CommitEntryMsg}
	param := MessageRequest{Message: c.CommitEntryMsg}
	req := primitives.NewJSON2Request("commit-entry", 1, param)

	_, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	//log.Print( jsonResp.Result.(*RevealEntryResponse).Message)

	// this is the blank '200 ok' that is returned for V1
	returnV1Msg(writer, "", true)

}

func HandleRevealEntry(writer http.ResponseWriter, request *http.Request) {
	type revealentry struct {
		Entry string
	}

	e := new(revealentry)
	if p, err := ioutil.ReadAll(request.Body); err != nil {
		handleV1Error(writer, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, e); err != nil {
			handleV1Error(writer, NewInvalidParamsError())
			return
		}
	}

	param := EntryRequest{Entry: e.Entry}
	req := primitives.NewJSON2Request("reveal-entry", 1, param)

	_, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	//log.Print(jsonResp.Result.(*RevealEntryResponse).Message)
	returnV1Msg(writer, "", true)

}

func HandleDirectoryBlockHead(writer http.ResponseWriter, request *http.Request) {
	req := primitives.NewJSON2Request("directory-block-head", 1, nil)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	tmp, err := json.Marshal(jsonResp.Result)
	resp := string(tmp)
	if err != nil {
		resp = "{\"KeyMR\",0}"
		returnV1Msg(writer, resp, true)
	}

	resp = strings.Replace(resp, "keymr", "KeyMR", -1)

	returnV1Msg(writer, resp, true)
}

func HandleGetRaw(writer http.ResponseWriter, request *http.Request) {
	param := extractURLHashParam(request)
	req := primitives.NewJSON2Request("raw-data", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	returnV1(writer, jsonResp, jsonError)
}

func HandleGetReceipt(writer http.ResponseWriter, request *http.Request) {
	param := extractURLHashParam(request)
	req := primitives.NewJSON2Request("receipt", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	returnV1(writer, jsonResp, jsonError)
}

func HandleDirectoryBlock(writer http.ResponseWriter, request *http.Request) {
	param := extractURLKeyMRParam(request)
	req := primitives.NewJSON2Request("directory-block", 1, param)
	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}

	type DBlock struct {
		Header struct {
			PrevBlockKeyMR string
			SequenceNumber int64
			Timestamp      int64
		}
		EntryBlockList []EBlockAddr
	}
	d := new(DBlock)

	d.Header.PrevBlockKeyMR = jsonResp.Result.(*DirectoryBlockResponse).Header.PrevBlockKeyMR
	d.Header.SequenceNumber = jsonResp.Result.(*DirectoryBlockResponse).Header.SequenceNumber
	d.Header.Timestamp = jsonResp.Result.(*DirectoryBlockResponse).Header.Timestamp
	d.EntryBlockList = jsonResp.Result.(*DirectoryBlockResponse).EntryBlockList

	// conflict if I use local structs.  using a string replace on the structs that would be pointer handled (*DirectoryBlockResponse)
	bResp, err := json.Marshal(d)
	if err != nil {
		returnMsg(writer, d, true)
	}
	resp := string(bResp)
	resp = strings.Replace(resp, "{\"chainid\"", "{\"ChainID\"", -1)
	resp = strings.Replace(resp, ",\"keymr\":", ",\"KeyMR\":", -1)

	returnV1Msg(writer, resp, true)
}

func HandleDirectoryBlockHeight(writer http.ResponseWriter, request *http.Request) {
	req := primitives.NewJSON2Request("heights", 1, nil)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}

	resp := "{\"Height\":0}"

	// get the HeightsResponse from the return object
	p, err := json.Marshal(jsonResp.Result)
	if err != nil {
		returnV1Msg(writer, resp, true)
	}
	h := new(HeightsResponse)
	if err := json.Unmarshal(p, h); err != nil {
		returnV1Msg(writer, resp, true)
	}

	// return just the DirectoryBlockHeight from the HeightsResponse
	resp = fmt.Sprintf("{\"Height\":%d}", h.DirectoryBlockHeight)
	returnV1Msg(writer, resp, true)
}

func HandleEntryBlock(writer http.ResponseWriter, request *http.Request) {
	param := extractURLKeyMRParam(request)
	req := primitives.NewJSON2Request("entry-block", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}

	type EBlock struct {
		Header struct {
			BlockSequenceNumber int64
			ChainID             string
			PrevKeyMR           string
			Timestamp           int64
			DBHeight            int64
		}
		EntryList []EntryAddr
	}
	d := new(EBlock)

	d.Header.BlockSequenceNumber = jsonResp.Result.(*EntryBlockResponse).Header.BlockSequenceNumber
	d.Header.ChainID = jsonResp.Result.(*EntryBlockResponse).Header.ChainID
	d.Header.PrevKeyMR = jsonResp.Result.(*EntryBlockResponse).Header.PrevKeyMR
	d.Header.Timestamp = jsonResp.Result.(*EntryBlockResponse).Header.Timestamp
	d.Header.DBHeight = jsonResp.Result.(*EntryBlockResponse).Header.DBHeight
	d.EntryList = jsonResp.Result.(*EntryBlockResponse).EntryList

	returnMsg(writer, d, true)
}

func HandleEntry(writer http.ResponseWriter, request *http.Request) {
	param := extractURLHashParam(request)
	req := primitives.NewJSON2Request("entry", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	d := new(EntryStruct)

	d.ChainID = jsonResp.Result.(*EntryResponse).ChainID
	d.Content = jsonResp.Result.(*EntryResponse).Content
	d.ExtIDs = jsonResp.Result.(*EntryResponse).ExtIDs

	returnMsg(writer, d, true)
}

func HandleChainHead(writer http.ResponseWriter, request *http.Request) {
	param := extractURLChainIDParam(request)
	req := primitives.NewJSON2Request("chain-head", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return

	}
	// restatement of chead from structs file
	// v1 doesn't like the lcase in v2'
	type CHead struct {
		ChainHead string
	}

	d := new(CHead)
	d.ChainHead = jsonResp.Result.(*ChainHeadResponse).ChainHead
	returnMsg(writer, d, true)
}

func HandleEntryCreditBalance(writer http.ResponseWriter, request *http.Request) {
	type x struct {
		Response string
		Success  bool
	}

	param := extractURLAddressParam(request)
	req := primitives.NewJSON2Request("entry-credit-balance", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}

	t := new(x)
	t.Response = fmt.Sprint(jsonResp.Result.(*EntryCreditBalanceResponse).Balance)
	t.Success = true
	returnMsg(writer, t, true)
}

func HandleGetFee(writer http.ResponseWriter, request *http.Request) {
	req := primitives.NewJSON2Request("entry-credit-rate", 1, nil)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	type x struct{ Fee int64 }
	d := new(x)

	d.Fee = int64(jsonResp.Result.(*EntryCreditRateResponse).Rate)

	returnMsg(writer, d, true)
}

func HandleFactoidSubmit(writer http.ResponseWriter, request *http.Request) {
	type x struct {
		Response string
		Success  bool
	}

	type transaction struct{ Transaction string }
	t := new(transaction)

	var p []byte
	var err error
	if p, err = ioutil.ReadAll(request.Body); err != nil {
		handleV1Error(writer, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, t); err != nil {
			handleV1Error(writer, NewInvalidParamsError())
			return
		}
	}

	param := TransactionRequest{Transaction: t.Transaction}
	req := primitives.NewJSON2Request("factoid-submit", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	r := new(x)
	r.Response = jsonResp.Result.(*FactoidSubmitResponse).Message
	r.Success = true
	returnMsg(writer, r, true)
}

func HandleFactoidBalance(writer http.ResponseWriter, request *http.Request) {
	type x struct {
		Response string
		Success  bool
	}

	param := extractURLAddressParam(request)
	req := primitives.NewJSON2Request("factoid-balance", 1, param)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}

	r := new(x)
	r.Response = fmt.Sprint(jsonResp.Result.(*FactoidBalanceResponse).Balance)
	r.Success = true
	returnMsg(writer, r, true)

}

func HandleProperties(writer http.ResponseWriter, request *http.Request) {
	req := primitives.NewJSON2Request("properties", 1, nil)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	type x struct {
		Protocol_Version string
		Factomd_Version  string
	}
	d := new(x)
	d.Factomd_Version = jsonResp.Result.(*PropertiesResponse).FactomdVersion
	d.Protocol_Version = "0.0.0.0" // meaningless after v1
	returnMsg(writer, d, true)
}

func HandleHeights(writer http.ResponseWriter, request *http.Request) {
	req := primitives.NewJSON2Request("heights", 1, nil)

	jsonResp, jsonError := HandleV2Request(writer, request, req)
	if jsonError != nil {
		returnV1(writer, nil, jsonError)
		return
	}
	returnMsg(writer, jsonResp.Result, true)
}
