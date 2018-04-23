// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/FactomProject/factomd/common/primitives"
)

type StateSaverStruct struct {
	FastBoot         bool
	FastBootLocation string

	TmpState []byte
	Mutex    sync.Mutex
	Stop     bool
}

//To be increased whenever the data being saved changes from the last verion
const version = 8

func (sss *StateSaverStruct) StopSaving() {
	sss.Mutex.Lock()
	defer sss.Mutex.Unlock()
	sss.Stop = true
}

func (sss *StateSaverStruct) SaveDBStateList(ss *DBStateList, networkName string) error {
	//For now, to file. Later - to DB
	if sss.Stop == true {
		return nil
	}
	sss.Mutex.Lock()
	defer sss.Mutex.Unlock()

	//Don't save States after the server has booted - it might start it in a wrong state
	if ss.State.DBFinished == true {
		return nil
	}

	//Save only every 1000 states
	if ss.GetHighestSavedBlk()%1000 != 0 || ss.GetHighestSavedBlk() < 1000 {
		return nil
	}

	//Actually save data from previous cached state to prevent dealing with rollbacks
	if len(sss.TmpState) > 0 {
		err := SaveToFile(sss.TmpState, NetworkIDToFilename(networkName, sss.FastBootLocation))
		if err != nil {
			return err
		}
	}

	//Marshal state for future saving
	b, err := ss.MarshalBinary()
	if err != nil {
		return err
	}
	//adding an integrity check
	h := primitives.Sha(b)
	b = append(h.Bytes(), b...)
	sss.TmpState = b

	return nil
}

func (sss *StateSaverStruct) DeleteSaveState(networkName string) error {
	return DeleteFile(NetworkIDToFilename(networkName, sss.FastBootLocation))
}

func (sss *StateSaverStruct) LoadDBStateList(ss *DBStateList, networkName string) error {
	b, err := LoadFromFile(NetworkIDToFilename(networkName, sss.FastBootLocation))
	if err != nil {
		return nil
	}
	if b == nil {
		return nil
	}
	h := primitives.NewZeroHash()
	b, err = h.UnmarshalBinaryData(b)
	if err != nil {
		return nil
	}
	h2 := primitives.Sha(b)
	if h.IsSameAs(h2) == false {
		fmt.Printf("LoadDBStateList - Integrity hashes do not match!")
		return nil
		//return fmt.Errorf("Integrity hashes do not match")
	}

	return ss.UnmarshalBinary(b)
}

func NetworkIDToFilename(networkName string, fileLocation string) string {
	file := fmt.Sprintf("FastBoot_%s_v%v.db", networkName, version)
	if fileLocation != "" {
		return fmt.Sprintf("%v/%v", fileLocation, file)
	}
	return file
}

func SaveToFile(b []byte, filename string) error {
	err := ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromFile(filename string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func DeleteFile(filename string) error {
	return os.Remove(filename)
}
