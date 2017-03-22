// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"io/ioutil"
)

var tmpState []byte

func SaveDBStateList(ss *DBStateList) error {
	//For now, to file. Later - to DB

	//Save only every 1000 states
	if ss.GetHighestSavedBlk()%1000 != 0 || ss.GetHighestSavedBlk() < 1000 {
		return nil
	}

	//Actually save data from previous cached state to prevent dealing with rollbacks
	if len(tmpState) > 0 {
		err := SaveToFile(tmpState)
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

func LoadDBStateList(ss *DBStateList) error {
	b, err := LoadFromFile()
	if err != nil {
		return nil
	}
	if b == nil {
		return nil
	}

	return ss.UnmarshalBinary(b)
}

func SaveTheState(ss *SaveState) error {
	//For now, to file. Later - to DB

	//Save only every 1000 states
	if ss.DBHeight%1000 != 0 || ss.DBHeight < 1000 {
		return nil
	}

	//Actually save data from previous cached state to prevent dealing with rollbacks
	if len(tmpState) > 0 {
		err := SaveToFile(tmpState)
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

func LoadState(ss *SaveState) error {
	b, err := LoadFromFile()
	if err != nil {
		return nil
	}
	if b == nil {
		return nil
	}

	return ss.UnmarshalBinary(b)
}

func SaveToFile(b []byte) error {
	err := ioutil.WriteFile("ss.test", b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromFile() ([]byte, error) {
	b, err := ioutil.ReadFile("ss.test")
	if err != nil {
		return nil, err
	}
	return b, nil
}
