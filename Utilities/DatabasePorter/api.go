package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/FactomProject/factomd/common/directoryBlock"
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

func GetDBlockHead() (string, error) {
	return "47ecd1198c7888e2b6236dbbec6a90d36f5ff23c69d46d1369b2d11ef8c74d38", nil

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
