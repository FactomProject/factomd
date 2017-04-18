// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"io/ioutil"
)

var tmpState []byte

//To be increased whenever the data being saved changes from the last verion
const version = 2

func SaveDBStateList(ss *DBStateList, networkName string) error {
	//For now, to file. Later - to DB

	//Don't save States after the server has booted - it might start it in a wrong state
	if ss.State.DBFinished == true {
		return nil
	}

	//Save only every 1000 states
	if ss.GetHighestSavedBlk()%1000 != 0 || ss.GetHighestSavedBlk() < 1000 {
		return nil
	}

	//Actually save data from previous cached state to prevent dealing with rollbacks
	if len(tmpState) > 0 {
		err := SaveToFile(tmpState, NetworkIDToFilename(networkName))
		if err != nil {
			return err
		}
	}

	//Marshal state for future saving
	b, err := ss.MarshalBinary()
	if err != nil {
		return err
	}
	tmpState = b

	return nil
}

func LoadDBStateList(ss *DBStateList, networkName string) error {
	b, err := LoadFromFile(NetworkIDToFilename(networkName))
	if err != nil {
		return nil
	}
	if b == nil {
		return nil
	}

	return ss.UnmarshalBinary(b)
}

/*
func SaveTheState(ss *SaveState, networkName string) error {
	//For now, to file. Later - to DB

	//Save only every 1000 states
	if ss.DBHeight%1000 != 0 || ss.DBHeight < 1000 {
		return nil
	}

	//Actually save data from previous cached state to prevent dealing with rollbacks
	if len(tmpState) > 0 {
		err := SaveToFile(tmpState, NetworkIDToFilename(networkName))
		if err != nil {
			return err
		}
	}

	//Marshal state for future saving
	b, err := ss.MarshalBinary()
	if err != nil {
		return err
	}
	tmpState = b

	return nil
}
*/

func NetworkIDToFilename(networkName string) string {
	return fmt.Sprintf("FastBoot_%s_v%v.db", networkName, version)
}

/*
func LoadState(ss *SaveState, networkName string) error {
	b, err := LoadFromFile(NetworkIDToFilename(networkName))
	if err != nil {
		return nil
	}
	if b == nil {
		return nil
	}

	return ss.UnmarshalBinary(b)
}
*/

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
