// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/primitives"
)

type StateSaverStruct struct {
	FastBoot         bool
	FastBootLocation string

	TmpDBHt  uint32
	TmpState []byte
	Mutex    sync.Mutex
	Stop     bool
	Saved    bool
}

func (sss *StateSaverStruct) StopSaving() {
	sss.Mutex.Lock()
	defer sss.Mutex.Unlock()
	sss.Stop = true
}

func (sss *StateSaverStruct) SaveDBStateList(s *State, ss *DBStateList, networkName string) error {
	//For now, to file. Later - to DB
	if sss.Stop == true {
		return nil // if we have closed the database then don't save
	}

	hsb := int(ss.GetHighestSavedBlk())
	//Save only every FastSaveRate states

	if hsb%ss.State.FastSaveRate != 0 || hsb < ss.State.FastSaveRate {
		return nil
	}

	sss.Mutex.Lock()
	defer sss.Mutex.Unlock()
	//Actually save data from previous cached state to prevent dealing with rollbacks
	// Save the N block old state and then make a new savestate for the next save
	if len(sss.TmpState) > 0 {
		if !sss.Saved {
			filename := NetworkIDToFilename(networkName, sss.FastBootLocation)
			s.LogPrintf("executeMsg", "%d-:-%d %20s Saving %s for dbht %d", s.LLeaderHeight, s.CurrentMinute, s.FactomNodeName, filename, sss.TmpDBHt)
			err := SaveToFile(s, sss.TmpDBHt, sss.TmpState, filename)
			if err != nil {
				fmt.Fprintln(os.Stderr, "SaveState SaveToFile Failed", err)
				return err
			}
			sss.Saved = true
		} else {
			panic("Attempt to save already saved savestate")
		}
	}

	//Marshal state for future saving
	b, err := ss.MarshalBinary()
	if err != nil {
		fmt.Fprintln(os.Stderr, "SaveState MarshalBinary Failed", err)
		return err
	}
	//adding an integrity check
	h := primitives.Sha(b)
	b = append(h.Bytes(), b...)
	sss.TmpState = b
	sss.TmpDBHt = ss.State.LLeaderHeight

	return nil
}

func (sss *StateSaverStruct) DeleteSaveState(networkName string) error {
	return DeleteFile(NetworkIDToFilename(networkName, sss.FastBootLocation))
}

func (sss *StateSaverStruct) LoadDBStateList(s *State, statelist *DBStateList, networkName string) error {
	filename := NetworkIDToFilename(networkName, sss.FastBootLocation)
	fmt.Println(statelist.State.FactomNodeName, "Loading from", filename)
	b, err := LoadFromFile(s, filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "LoadDBStateList error:", err)
		return err
	}
	if b == nil {
		fmt.Fprintln(os.Stderr, "LoadDBStateList LoadFromFile returned nil")
		return errors.New("failed to load from file")
	}
	h := primitives.NewZeroHash()
	b, err = h.UnmarshalBinaryData(b)
	if err != nil {
		return err
	}
	h2 := primitives.Sha(b)
	if h.IsSameAs(h2) == false {
		fmt.Fprintf(os.Stderr, "LoadDBStateList - Integrity hashes do not match!")
		return errors.New("fastboot file does not match its hash")
		//return fmt.Errorf("Integrity hashes do not match")
	}

	statelist.UnmarshalBinary(b)
	var i int
	for i = len(statelist.DBStates) - 1; i >= 0; i-- {
		if statelist.DBStates[i].SaveStruct != nil {
			break
		}
	}
	statelist.DBStates[i].SaveStruct.RestoreFactomdState(statelist.State)

	return nil
}

func NetworkIDToFilename(networkName string, fileLocation string) string {
	file := fmt.Sprintf("FastBoot_%s_v%v.db", networkName, constants.SaveStateVersion)
	if fileLocation != "" {
		// Trim optional trailing / from file path
		i := len(fileLocation) - 1
		if fileLocation[i] == byte('/') {
			fileLocation = fileLocation[:i]
		}
		return fmt.Sprintf("%v/%v", fileLocation, file)
	}
	return file
}

func SaveToFile(s *State, dbht uint32, b []byte, filename string) error {
	fmt.Fprintf(os.Stderr, "%20s Saving %s for dbht %d\n", s.FactomNodeName, filename, dbht)
	err := ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%20s Saving FailrueError: %v\n", s.FactomNodeName, err)
		return err
	}
	return nil
}

func LoadFromFile(s *State, filename string) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "%20s Load state from %s\n", s.FactomNodeName, filename)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%20s LoadFromFile error: %v\n", s.FactomNodeName, err)
		return nil, err
	}
	return b, nil
}

func DeleteFile(filename string) error {
	return os.Remove(filename)
}
