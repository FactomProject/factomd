package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
)

//var server string = "localhost:8088" //Localhost
//var server string = "52.17.183.121:8088" //TestNet
var server string = "52.18.72.212:8088" //MainNet

type DBlockHead struct {
	KeyMR string
}

func GetDBlock(keymr string) (interfaces.IDirectoryBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	dblock, err := directoryBlock.UnmarshalDBlock(raw)
	if err != nil {
		return nil, err
	}
	return dblock, nil
}

func GetABlock(keymr string) (interfaces.IAdminBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	block, err := adminBlock.UnmarshalABlock(raw)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetECBlock(keymr string) (interfaces.IEntryCreditBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	block, err := entryCreditBlock.UnmarshalECBlock(raw)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetFBlock(keymr string) (interfaces.IFBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	block, err := block.UnmarshalFBlock(raw)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetEBlock(keymr string) (interfaces.IEntryBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	block, err := entryBlock.UnmarshalEBlock(raw)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetEntry(hash string) (interfaces.IEBEntry, error) {
	raw, err := GetRaw(hash)
	if err != nil {
		return nil, err
	}
	entry, err := entryBlock.UnmarshalEntry(raw)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func GetDBlockHead() (string, error) {
	return "3a5ec711a1dc1c6e463b0c0344560f830eb0b56e42def141cb423b0d8487a1dc", nil

	resp, err := http.Get(
		fmt.Sprintf("http://%s/v1/directory-block-head/", server))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf(string(body))
	}

	d := new(DBlockHead)
	json.Unmarshal(body, d)

	return d.KeyMR, nil
}

type Data struct {
	Data string
}

func GetRaw(keymr string) ([]byte, error) {
	resp, err := http.Get(
		fmt.Sprintf("http://%s/v1/get-raw-data/%s", server, keymr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(string(body))
	}

	d := new(Data)
	if err := json.Unmarshal(body, d); err != nil {
		return nil, err
	}

	raw, err := hex.DecodeString(d.Data)
	if err != nil {
		return nil, err
	}

	return raw, nil
}
