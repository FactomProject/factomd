// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hoisie/web"
)

func handleCommitChain(ctx *web.Context, name string) {
	type commit struct {
		CommitChainMsg string
	}

	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Println("Could not read from http request:", err)
		ctx.WriteHeader(httpBad)
		return
	}
	signed := factoidState.GetWallet().SignCommit([]byte(name), data)

	com := new(commit)
	com.CommitChainMsg = hex.EncodeToString(signed)
	j, err := json.Marshal(com)
	if err != nil {
		fmt.Println("Could not create json post:", err)
		ctx.WriteHeader(httpBad)
		return
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/commit-chain/", ipaddressFD+portNumberFD),
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		fmt.Println("Could not post to server:", err)
		ctx.WriteHeader(httpBad)
		return
	}
	resp.Body.Close()
}

func handleCommitEntry(ctx *web.Context, name string) {
	type commit struct {
		CommitEntryMsg string
	}

	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Println("Could not read from http request:", err)
		ctx.WriteHeader(httpBad)
		return
	}
	signed := factoidState.GetWallet().SignCommit([]byte(name), data)

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