// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"github.com/hoisie/web"
)

func HandleCommitChain(ctx *web.Context, name string) {
	type walletcommit struct {
		Message string
	}

	type commit struct {
		CommitChainMsg string
	}

	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Println("Could not read from http request:", err)
		ctx.WriteHeader(httpBad)
		return
	}
	in := new(walletcommit)
	json.Unmarshal(data, in)
	msg, err := hex.DecodeString(in.Message)
	if err != nil {
		fmt.Println("Could not decode message:", err)
		ctx.WriteHeader(httpBad)
		return
	}

	we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
	signed := factoidState.GetWallet().SignCommit(we.(wallet.IWalletEntry), msg)

	com := new(commit)
	com.CommitChainMsg = hex.EncodeToString(signed)
	j, err := json.Marshal(com)
	if err != nil {
		fmt.Println("Could not create json post:", err)
		ctx.WriteHeader(httpBad)
		return
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/commit-chain", ipaddressFD+portNumberFD),
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		fmt.Println("Could not post to server:", err)
		ctx.WriteHeader(httpBad)
		return
	}
	resp.Body.Close()
}

func HandleCommitEntry(ctx *web.Context, name string) {
	type walletcommit struct {
		Message string
	}

	type commit struct {
		CommitEntryMsg string
	}

	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Println("Could not read from http request:", err)
		ctx.WriteHeader(httpBad)
		return
	}
	in := new(walletcommit)
	json.Unmarshal(data, in)
	msg, err := hex.DecodeString(in.Message)
	if err != nil {
		fmt.Println("Could not decode message:", err)
		ctx.WriteHeader(httpBad)
		return
	}

	we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
	signed := factoidState.GetWallet().SignCommit(we.(wallet.IWalletEntry), msg)

	com := new(commit)
	com.CommitEntryMsg = hex.EncodeToString(signed)
	j, err := json.Marshal(com)
	if err != nil {
		fmt.Println("Could not create json post:", err)
		ctx.WriteHeader(httpBad)
		return
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/commit-entry/", ipaddressFD+portNumberFD),
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		fmt.Println("Could not post to server:", err)
		ctx.WriteHeader(httpBad)
		return
	}
	resp.Body.Close()
}
