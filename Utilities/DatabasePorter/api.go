package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/PaulSnow/factom2d/common/adminBlock"
	"github.com/PaulSnow/factom2d/common/directoryBlock"
	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/common/interfaces"
)

var server string = "localhost:8088" //Localhost
//var server string = "52.17.183.121:8088" //TestNet
//var server string = "52.18.72.212:8088" //MainNet

type DBlockHead struct {
	KeyMR string
}

func GetDBlock(keymr string) (interfaces.IDirectoryBlock, error) {
	for i := 0; i < 100; i++ {
		raw, err := GetRaw(keymr)
		if err != nil {
			continue
		}
		dblock, err := directoryBlock.UnmarshalDBlock(raw)
		if err != nil {
			continue
		}
		return dblock, nil
	}
	panic("Failed 100 times to get the data " + keymr)
	return nil, nil
}

func GetABlock(keymr string) (interfaces.IAdminBlock, error) {
	for i := 0; i < 100; i++ {
		raw, err := GetRaw(keymr)
		if err != nil {
			continue
		}

		block, err := adminBlock.UnmarshalABlock(raw)
		if err != nil {
			continue
		}
		return block, nil
	}
	panic("Failed 100 times to get the data " + keymr)
	return nil, nil
}

func GetECBlock(keymr string) (interfaces.IEntryCreditBlock, error) {
	for i := 0; i < 100; i++ {
		raw, err := GetRaw(keymr)
		if err != nil {
			continue
		}
		block, err := entryCreditBlock.UnmarshalECBlock(raw)
		if err != nil {
			continue
		}
		return block, nil
	}
	panic("Failed 100 times to get the data " + keymr)
	return nil, nil
}

func GetFBlock(keymr string) (interfaces.IFBlock, error) {
	for i := 0; i < 100; i++ {
		raw, err := GetRaw(keymr)
		if err != nil {
			continue
		}
		block, err := factoid.UnmarshalFBlock(raw)
		if err != nil {
			continue
		}
		return block, nil
	}
	panic("Failed 100 times to get the data " + keymr)
	return nil, nil
}

func GetEBlock(keymr string) (interfaces.IEntryBlock, error) {
	for i := 0; i < 100; i++ {
		raw, err := GetRaw(keymr)
		if err != nil {
			continue
		}
		block, err := entryBlock.UnmarshalEBlock(raw)
		if err != nil {
			continue
		}
		return block, nil
	}
	panic("Failed 100 times to get the data " + keymr)
	return nil, nil
}

func GetEntry(hash string) (interfaces.IEBEntry, error) {
	for i := 0; i < 10; i++ {
		raw, err := GetRaw(hash)
		if err != nil {
			fmt.Printf("got error %s\n", err)
			fmt.Printf("called getraw with %s\n", hash)
			fmt.Printf("got result %s\n", raw)

			continue
		}
		entry, err := entryBlock.UnmarshalEntry(raw)
		for err != nil { //just keep trying until it doesn't give an error
			fmt.Printf("got error %s\n", err)
			fmt.Printf("called entryBlock.UnmarshalEntry with %s\n", raw)
			fmt.Printf("got result %s\n", entry)
			//if we get an error like EOF, get the thing again after a short wait
			time.Sleep(20000 * time.Millisecond)
			raw, err = GetRaw(hash)
			if err != nil {
				continue
			}
			entry, err = entryBlock.UnmarshalEntry(raw)
		}
		return entry, nil
	}
	//panic("Failed 100 times to get the data " + hash)
	return nil, nil
}

func GetDBlockHead() (string, error) {
	//return "3a5ec711a1dc1c6e463b0c0344560f830eb0b56e42def141cb423b0d8487a1dc", nil //10
	//return "cde346e7ed87957edfd68c432c984f35596f29c7d23de6f279351cddecd5dc66", nil //100
	//return "d13472838f0156a8773d78af137ca507c91caf7bf3b73124d6b09ebb0a98e4d9", nil //200

	for i := 0; i < 100; i++ {
		resp, err := http.Get(
			fmt.Sprintf("http://%s/v1/directory-block-head/", server))
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		if resp.StatusCode != 200 {
			continue
		}

		d := new(DBlockHead)
		json.Unmarshal(body, d)

		return d.KeyMR, nil
	}
	panic("Failed 100 times to get the DBlock Head")
	return "", nil
}

type Data struct {
	Data string
}

func GetRaw(keymr string) ([]byte, error) {
	for i := 0; i < 100; i++ {
		resp, err := http.Get(fmt.Sprintf("http://%s/v1/get-raw-data/%s", server, keymr))
		for err != nil {
			//if the http code gave an error, give a little time and try again before panicking.
			fmt.Printf("got error %s, waiting 20 seconds\n", err)
			time.Sleep(20000 * time.Millisecond)
			resp, err = http.Get(fmt.Sprintf("http://%s/v1/get-raw-data/%s", server, keymr))
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		for err != nil {
			//if the io reader code gave an error, give a little time and try again before panicking.
			fmt.Printf("got error %s, waiting 20 seconds\n", err)
			time.Sleep(20000 * time.Millisecond)
			body, err = ioutil.ReadAll(resp.Body)
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf(string(body))
		}

		d := new(Data)
		if err := json.Unmarshal(body, d); err != nil {
			continue
		}

		raw, err := hex.DecodeString(d.Data)
		if err != nil {
			continue
		}

		return raw, nil
	}
	panic("Failed 100 times to get the data " + keymr)
	return nil, nil
}
