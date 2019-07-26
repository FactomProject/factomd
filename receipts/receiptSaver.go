// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package receipts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var DataStorePath string = "./receipts"

func FileNotExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true
	}
	return err != nil
}

func Save(receipt *Receipt) error {
	data, err := receipt.JSONByte()
	if err != nil {
		return err
	}

	var out bytes.Buffer
	json.Indent(&out, data, "", "\t")
	data = out.Next(out.Len())

	entryID := receipt.Entry.EntryHash
	dir := DataStorePath // + entryID
	if FileNotExists(dir) {
		err := os.MkdirAll(dir, 0777)
		if err == nil {
			fmt.Println("Created directory " + dir)
		} else {
			return err
		}
	}

	fmt.Printf("Saving %v\n", fmt.Sprintf(dir+"/storeJSON.%v.block", entryID))

	err = ioutil.WriteFile(fmt.Sprintf(dir+"/storeJSON.%v.block", entryID), data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func ExportEntryReceipt(entryID string, dbo interfaces.DBOverlaySimple) error {
	h, err := primitives.NewShaHashFromStr(entryID)
	if err != nil {
		return err
	}
	receipt, err := CreateFullReceipt(dbo, h, false)
	if err != nil {
		return err
	}
	return Save(receipt)
}

func ExportAllEntryReceipts(dbo interfaces.DBOverlay) error {
	entryIDs, err := dbo.FetchAllEntryIDs()
	if err != nil {
		return err
	}
	for i, entryID := range entryIDs {
		err = ExportEntryReceipt(entryID.String(), dbo)
		if err != nil {
			if err.Error() != "dirBlockInfo not found" {
				return err
			} else {
				fmt.Printf("dirBlockInfo not found for entry %v/%v - %v\n", i, len(entryIDs), entryID)
			}
		}
	}
	return nil
}
