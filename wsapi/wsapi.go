// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FactomProject/btcutil/certs"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/web"
)

const (
	httpBad = 400
)

var Servers map[int]*web.Server
var ServersMutex sync.Mutex

func Start(state interfaces.IState) {
	RegisterPrometheus()
	var server *web.Server

	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	if Servers == nil {
		Servers = make(map[int]*web.Server)
	}

	rpcUser := state.GetRpcUser()
	rpcPass := state.GetRpcPass()
	h := sha256.New()
	h.Write(httpBasicAuth(rpcUser, rpcPass))
	state.SetRpcAuthHash(h.Sum(nil)) //set this in the beginning to prevent timing attacks

	if Servers[state.GetPort()] == nil {
		server = web.NewServer()

		server.Logger.SetOutput(ioutil.Discard)
		server.Config.CorsDomains = state.GetCorsDomains()

		Servers[state.GetPort()] = server
		server.Env["state"] = state

		server.Post("/v1/factoid-submit/?", HandleFactoidSubmit)
		server.Post("/v1/commit-chain/?", HandleCommitChain)
		server.Post("/v1/reveal-chain/?", HandleRevealChain)
		server.Post("/v1/commit-entry/?", HandleCommitEntry)
		server.Post("/v1/reveal-entry/?", HandleRevealEntry)
		server.Get("/v1/directory-block-head/?", HandleDirectoryBlockHead)
		server.Get("/v1/get-raw-data/([^/]+)", HandleGetRaw)
		server.Get("/v1/get-receipt/([^/]+)", HandleGetReceipt)
		server.Get("/v1/directory-block-by-keymr/([^/]+)", HandleDirectoryBlock)
		server.Get("/v1/directory-block-height/?", HandleDirectoryBlockHeight)
		server.Get("/v1/entry-block-by-keymr/([^/]+)", HandleEntryBlock)
		server.Get("/v1/entry-by-hash/([^/]+)", HandleEntry)
		server.Get("/v1/chain-head/([^/]+)", HandleChainHead)
		server.Get("/v1/entry-credit-balance/([^/]+)", HandleEntryCreditBalance)
		server.Get("/v1/factoid-balance/([^/]+)", HandleFactoidBalance)
		server.Get("/v1/factoid-get-fee/", HandleGetFee)
		server.Get("/v1/properties/", HandleProperties)
		server.Get("/v1/heights/", HandleHeights)

		server.Get("/v1/dblock-by-height/([^/]+)", HandleDBlockByHeight)
		server.Get("/v1/ecblock-by-height/([^/]+)", HandleECBlockByHeight)
		server.Get("/v1/fblock-by-height/([^/]+)", HandleFBlockByHeight)
		server.Get("/v1/ablock-by-height/([^/]+)", HandleABlockByHeight)

		server.Post("/v2", HandleV2)
		server.Get("/v2", HandleV2)

		server.Get("/status", HandleStatus)

		// start the debugging api if we are not on the main network
		if state.GetNetworkName() != "MAIN" {
			server.Post("/debug", HandleDebug)
			server.Get("/debug", HandleDebug)
		}

		tlsIsEnabled, tlsPrivate, tlsPublic := state.GetTlsInfo()
		if tlsIsEnabled {
			log.Print("Starting encrypted API server")
			if !fileExists(tlsPrivate) && !fileExists(tlsPublic) {
				err := genCertPair(tlsPublic, tlsPrivate, state.GetFactomdLocations())
				if err != nil {
					panic(fmt.Sprintf("could not start encrypted API server with error: %v", err))
				}
			}
			keypair, err := tls.LoadX509KeyPair(tlsPublic, tlsPrivate)
			if err != nil {
				panic(fmt.Sprintf("could not create TLS keypair with error: %v", err))
			}
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{keypair},
				MinVersion:   tls.VersionTLS12,
			}
			go server.RunTLS(fmt.Sprintf(":%d", state.GetPort()), tlsConfig)

		} else {
			log.Print("Starting API server")
			go server.Run(fmt.Sprintf(":%d", state.GetPort()))
		}
	}
}

func SetState(state interfaces.IState) {
	wait := func() {
		ServersMutex.Lock()
		defer ServersMutex.Unlock()
		//todo: Should wait() instead of sleep but that requires plumbing a wait group....
		for Servers == nil && Servers[state.GetPort()] != nil && Servers[state.GetPort()].Env != nil {
			ServersMutex.Unlock()
			time.Sleep(10 * time.Millisecond)
			ServersMutex.Lock()
		}
		gp := state.GetPort()
		if Servers == nil {
			fmt.Println("Got here early need synchronization")
		}
		if Servers[gp] == nil {
			fmt.Println("Got here early need synchronization")
		}
		if Servers[gp].Env == nil {
			fmt.Println("Got here early need synchronization")
		}

		Servers[gp].Env["state"] = state
	}
	go wait()
}

func Stop(state interfaces.IState) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	Servers[state.GetPort()].Close()
}

func handleV1Error(ctx *web.Context, err *primitives.JSONError) {
	/*
		if err.Data != nil {
			data, ok := err.Data.(string)
			if ok == true {
		ctx.WriteHeader(httpBad)
				returnMsg(ctx, "", false)
				return
			}
		}
		ctx.WriteHeader(httpBad)
		returnMsg(ctx,"", false)
		return
	*/
	ctx.WriteHeader(httpBad)
	return
}

func returnV1(ctx *web.Context, jsonResp *primitives.JSON2Response, jsonError *primitives.JSONError) {
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

func HandleDBlockByHeight(ctx *web.Context, height string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	h, err := strconv.ParseInt(height, 0, 64)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}
	param := HeightRequest{Height: h}
	req := primitives.NewJSON2Request("dblock-by-height", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleECBlockByHeight(ctx *web.Context, height string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	h, err := strconv.ParseInt(height, 0, 64)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}
	param := HeightRequest{Height: h}
	req := primitives.NewJSON2Request("ecblock-by-height", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleFBlockByHeight(ctx *web.Context, height string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	h, err := strconv.ParseInt(height, 0, 64)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}
	param := HeightRequest{Height: h}
	req := primitives.NewJSON2Request("fblock-by-height", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleABlockByHeight(ctx *web.Context, height string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	h, err := strconv.ParseInt(height, 0, 64)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}
	param := HeightRequest{Height: h}
	req := primitives.NewJSON2Request("ablock-by-height", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleCommitChain(ctx *web.Context) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	type commitchain struct {
		CommitChainMsg string
	}
	c := new(commitchain)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, c); err != nil {
			handleV1Error(ctx, NewInvalidParamsError())
			return
		}
	}
	param := MessageRequest{Message: c.CommitChainMsg}
	req := primitives.NewJSON2Request("commit-chain", 1, param)
	_, jsonError := HandleV2Request(state, req)

	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	//log.Print(jsonResp.Result.(*RevealEntryResponse).Message)

	// this is the blank '200 ok' that is returned for V1
	returnV1Msg(ctx, "", false)

}

func HandleRevealChain(ctx *web.Context) {
	HandleRevealEntry(ctx)
}

func HandleCommitEntry(ctx *web.Context) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	type commitentry struct {
		CommitEntryMsg string
	}

	c := new(commitentry)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, c); err != nil {
			handleV1Error(ctx, NewInvalidParamsError())
			return
		}
	}

	//  v2 wants a MessageRequest instead of an EntryRequest
	//	param := EntryRequest{Entry: c.CommitEntryMsg}
	param := MessageRequest{Message: c.CommitEntryMsg}
	req := primitives.NewJSON2Request("commit-entry", 1, param)

	_, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	//log.Print( jsonResp.Result.(*RevealEntryResponse).Message)

	// this is the blank '200 ok' that is returned for V1
	returnV1Msg(ctx, "", true)

}

func HandleRevealEntry(ctx *web.Context) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	type revealentry struct {
		Entry string
	}

	e := new(revealentry)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, e); err != nil {
			handleV1Error(ctx, NewInvalidParamsError())
			return
		}
	}

	param := EntryRequest{Entry: e.Entry}
	req := primitives.NewJSON2Request("reveal-entry", 1, param)

	_, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	//log.Print(jsonResp.Result.(*RevealEntryResponse).Message)
	returnV1Msg(ctx, "", true)

}

func HandleDirectoryBlockHead(ctx *web.Context) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	req := primitives.NewJSON2Request("directory-block-head", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	tmp, err := json.Marshal(jsonResp.Result)
	resp := string(tmp)
	if err != nil {
		resp = "{\"KeyMR\",0}"
		returnV1Msg(ctx, resp, true)
	}

	resp = strings.Replace(resp, "keymr", "KeyMR", -1)

	returnV1Msg(ctx, resp, true)
}

func HandleGetRaw(ctx *web.Context, hashkey string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	param := HashRequest{Hash: hashkey}
	req := primitives.NewJSON2Request("raw-data", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleGetReceipt(ctx *web.Context, hashkey string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	param := HashRequest{Hash: hashkey}
	req := primitives.NewJSON2Request("receipt", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleDirectoryBlock(ctx *web.Context, hashkey string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	param := KeyMRRequest{KeyMR: hashkey}
	req := primitives.NewJSON2Request("directory-block", 1, param)
	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
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
		returnMsg(ctx, d, true)
	}
	resp := string(bResp)
	resp = strings.Replace(resp, "{\"chainid\"", "{\"ChainID\"", -1)
	resp = strings.Replace(resp, ",\"keymr\":", ",\"KeyMR\":", -1)

	returnV1Msg(ctx, resp, true)
}

func HandleDirectoryBlockHeight(ctx *web.Context) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	req := primitives.NewJSON2Request("heights", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	resp := "{\"Height\":0}"

	// get the HeightsResponse from the return object
	p, err := json.Marshal(jsonResp.Result)
	if err != nil {
		returnV1Msg(ctx, resp, true)
	}
	h := new(HeightsResponse)
	if err := json.Unmarshal(p, h); err != nil {
		returnV1Msg(ctx, resp, true)
	}

	// return just the DirectoryBlockHeight from the HeightsResponse
	resp = fmt.Sprintf("{\"Height\":%d}", h.DirectoryBlockHeight)
	returnV1Msg(ctx, resp, true)
}

func HandleEntryBlock(ctx *web.Context, hashkey string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	param := KeyMRRequest{KeyMR: hashkey}
	req := primitives.NewJSON2Request("entry-block", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
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

	returnMsg(ctx, d, true)
}

func HandleEntry(ctx *web.Context, hashkey string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	param := HashRequest{Hash: hashkey}
	req := primitives.NewJSON2Request("entry", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	d := new(EntryStruct)

	d.ChainID = jsonResp.Result.(*EntryResponse).ChainID
	d.Content = jsonResp.Result.(*EntryResponse).Content
	d.ExtIDs = jsonResp.Result.(*EntryResponse).ExtIDs

	returnMsg(ctx, d, true)
}

func HandleChainHead(ctx *web.Context, chainid string) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	param := ChainIDRequest{ChainID: chainid}
	req := primitives.NewJSON2Request("chain-head", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return

	}
	// restatement of chead from structs file
	// v1 doesn't like the lcase in v2'
	type CHead struct {
		ChainHead string
	}

	d := new(CHead)
	d.ChainHead = jsonResp.Result.(*ChainHeadResponse).ChainHead
	returnMsg(ctx, d, true)
}

func HandleEntryCreditBalance(ctx *web.Context, address string) {
	type x struct {
		Response string
		Success  bool
	}

	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	param := AddressRequest{Address: address}
	req := primitives.NewJSON2Request("entry-credit-balance", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	t := new(x)
	t.Response = fmt.Sprint(jsonResp.Result.(*EntryCreditBalanceResponse).Balance)
	t.Success = true
	returnMsg(ctx, t, true)
}

func HandleGetFee(ctx *web.Context) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	req := primitives.NewJSON2Request("entry-credit-rate", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	type x struct{ Fee int64 }
	d := new(x)

	d.Fee = int64(jsonResp.Result.(*EntryCreditRateResponse).Rate)

	returnMsg(ctx, d, true)
}

func HandleFactoidSubmit(ctx *web.Context) {
	type x struct {
		Response string
		Success  bool
	}

	type transaction struct{ Transaction string }
	t := new(transaction)

	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	var p []byte
	var err error
	if p, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, t); err != nil {
			handleV1Error(ctx, NewInvalidParamsError())
			return
		}
	}

	param := TransactionRequest{Transaction: t.Transaction}
	req := primitives.NewJSON2Request("factoid-submit", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	r := new(x)
	r.Response = jsonResp.Result.(*FactoidSubmitResponse).Message
	r.Success = true
	returnMsg(ctx, r, true)
}

func HandleFactoidBalance(ctx *web.Context, address string) {
	type x struct {
		Response string
		Success  bool
	}

	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	param := AddressRequest{Address: address}
	req := primitives.NewJSON2Request("factoid-balance", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	r := new(x)
	r.Response = fmt.Sprint(jsonResp.Result.(*FactoidBalanceResponse).Balance)
	r.Success = true
	returnMsg(ctx, r, true)

}

func HandleProperties(ctx *web.Context) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)

	if !checkHttpPasswordOkV1(state, ctx) {
		return
	}

	req := primitives.NewJSON2Request("properties", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	type x struct {
		Protocol_Version string
		Factomd_Version  string
	}
	d := new(x)
	d.Factomd_Version = jsonResp.Result.(*PropertiesResponse).FactomdVersion
	d.Protocol_Version = "0.0.0.0" // meaningless after v1
	returnMsg(ctx, d, true)
}

func HandleHeights(ctx *web.Context) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	state := ctx.Server.Env["state"].(interfaces.IState)
	req := primitives.NewJSON2Request("heights", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

type statusResponse struct {
	Version                      string
	NodeName                     string
	BootTime                     int64
	CurrentTime                  int64
	CurrentBlockStartTime        int64
	CurrentMinuteStartTime       int64
	CurrentMinute                int
	LeaderHeight                 uint32
	HighestSavedBlock            uint32
	HighestKnownBlock            uint32
	DBHeightComplete             uint32
	EntryBlockDBHeightComplete   uint32
	EntryBlockDBHeightProcessing uint32
	Syncing                      bool
	SyncingEOMs                  bool
	SyncingDBSigs                bool
	Running                      bool
	IgnoreDone                   bool
	Role                         string
}

func HandleStatus(ctx *web.Context) {
	ServersMutex.Lock()
	s := ctx.Server.Env["state"].(interfaces.IState)
	ServersMutex.Unlock()

	feds := s.GetFedServers(s.GetLLeaderHeight())
	audits := s.GetAuditServers(s.GetLLeaderHeight())
	role := "follower"
	foundRole := false
	for _, fed := range feds {
		if !foundRole && s.GetIdentityChainID().IsSameAs(fed.GetChainID()) {
			role = "leader"
			break
		}
	}
	for _, aud := range audits {
		if !foundRole && s.GetIdentityChainID().IsSameAs(aud.GetChainID()) {
			role = "audit"
		}
	}

	jsonResp := primitives.NewJSON2Response()
	jsonResp.ID = 0
	jsonResp.Result = statusResponse{
		s.GetFactomdVersion(),
		s.GetFactomNodeName(),
		s.GetBootTime(),
		s.GetCurrentTime(),
		s.GetCurrentBlockStartTime(),
		s.GetCurrentMinuteStartTime(),
		s.GetCurrentMinute(),
		s.GetLeaderHeight(),
		s.GetHighestSavedBlk(),
		s.GetHighestKnownBlock(),
		s.GetEntryBlockDBHeightComplete(),
		s.GetDBHeightComplete(),
		s.GetEntryBlockDBHeightProcessing(),
		s.IsSyncing(),
		s.IsSyncingEOMs(),
		s.IsSyncingDBSigs(),
		s.Running(),
		s.GetIgnoreDone(),
		role,
	}

	// REVIEW: should we only conditionally return status 200 if some precondition is met
	ctx.Write([]byte(jsonResp.String()))
}

/*********************************************************
 * Support Functions
 *********************************************************/

func returnMsg(ctx *web.Context, msg interface{}, success bool) {
	type rtn struct {
		Response interface{}
		Success  bool
	}
	/*str, ok:=msg.(string)
	if ok == false {
		var err error
		str, err = primitives.EncodeJSONString(msg)
		if err != nil {
			wsLog.Error(err)
			return
		}
	}*/
	r := msg

	if p, err := json.Marshal(r); err != nil {
		wsLog.Error(err)
		return
	} else {
		ctx.Write(p)
	}
}

func returnV1Msg(ctx *web.Context, msg string, success bool) {
	/* V1 requires call specific case changes that can't be handled with
	interfaces for example.  Block Height needs to return  height as the json item name
	in golang, lower case names are private so won't be returned.
	Deal with the responses in the call specific v1 handlers until they are depricated.
	*/
	bMsg := []byte(msg)
	ctx.Write(bMsg)

}

// httpBasicAuth returns the UTF-8 bytes of the HTTP Basic authentication
// string:
//
//   "Basic " + base64(username + ":" + password)
func httpBasicAuth(username, password string) []byte {
	const header = "Basic "
	base64 := base64.StdEncoding

	b64InputLen := len(username) + len(":") + len(password)
	b64Input := make([]byte, 0, b64InputLen)
	b64Input = append(b64Input, username...)
	b64Input = append(b64Input, ':')
	b64Input = append(b64Input, password...)

	output := make([]byte, len(header)+base64.EncodedLen(b64InputLen))
	copy(output, header)
	base64.Encode(output[len(header):], b64Input)
	return output
}

func checkAuthHeader(state interfaces.IState, r *http.Request) error {
	if "" == state.GetRpcUser() {
		//no username was specified in the config file or command line, meaning factomd API is open access
		return nil
	}

	authhdr := r.Header["Authorization"]
	if len(authhdr) == 0 {
		return errors.New("no auth")
	}

	correctAuth := state.GetRpcAuthHash()

	h := sha256.New()
	h.Write([]byte(authhdr[0]))
	presentedPassHash := h.Sum(nil)

	cmp := subtle.ConstantTimeCompare(presentedPassHash, correctAuth) //compare hashes because ConstantTimeCompare takes a constant time based on the slice size.  hashing gives a constant slice size.
	if cmp != 1 {
		return errors.New("bad auth")
	}
	return nil
}

func checkHttpPasswordOkV1(state interfaces.IState, ctx *web.Context) bool {
	if err := checkAuthHeader(state, ctx.Request); err != nil {
		remoteIP := ""
		remoteIP += strings.Split(ctx.Request.RemoteAddr, ":")[0]
		fmt.Printf("Unauthorized V1 API client connection attempt from %s\n", remoteIP)
		ctx.ResponseWriter.Header().Add("WWW-Authenticate", `Basic realm="factomd RPC"`)
		http.Error(ctx.ResponseWriter, "401 Unauthorized.", http.StatusUnauthorized)
		return false
	}
	return true
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func genCertPair(certFile string, keyFile string, extraAddress string) error {
	fmt.Println("Generating TLS certificates...")

	org := "factom autogenerated cert"
	validUntil := time.Now().Add(10 * 365 * 24 * time.Hour)

	var externalAddresses []string
	if extraAddress != "" {
		externalAddresses = strings.Split(extraAddress, ",")
		for _, i := range externalAddresses {
			fmt.Printf("adding %s to certificate\n", i)
		}
	}

	cert, key, err := certs.NewTLSCertPair(org, validUntil, externalAddresses)
	if err != nil {
		return err
	}

	// Write cert and key files.
	if err = ioutil.WriteFile(certFile, cert, 0666); err != nil {
		return err
	}
	if err = ioutil.WriteFile(keyFile, key, 0600); err != nil {
		os.Remove(certFile)
		return err
	}

	fmt.Println("Done generating TLS certificates")
	return nil
}
