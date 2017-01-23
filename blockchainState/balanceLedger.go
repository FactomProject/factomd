// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"bytes"
	"encoding/gob"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type BalanceLedger struct {
	Balances map[string]AddressBalance
	Deltas   []BalanceDelta
}

func (bs *BalanceLedger) Init() {
	if bs.Balances == nil {
		bs.Balances = map[string]AddressBalance{}
	}
}

func (bs *BalanceLedger) MarshalBinary() ([]byte, error) {
	b := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(b)
	err := enc.Encode(bs)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (bs *BalanceLedger) UnmarshalBinary(data []byte) error {
	bs.Init()
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	return dec.Decode(bs)
}

func (e *BalanceLedger) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *BalanceLedger) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *BalanceLedger) String() string {
	str, _ := e.JSONString()
	return str
}

func (bs *BalanceLedger) ProcessFBlock(fBlock interfaces.IFBlock) error {
	bs.Init()

	var delta BalanceDelta
	delta.Init()
	index := int(fBlock.GetDatabaseHeight())

	transactions := fBlock.GetTransactions()
	for _, tx := range transactions {
		ins := tx.GetInputs()
		for _, w := range ins {
			add := w.GetAddress().String()
			balance := -int64(w.GetAmount())

			_, exist := bs.Balances[add]
			if exist == false {
				//Ensure we have top-level balance
				bs.Balances[add] = AddressBalance{Address: add, Balance: 0, LastDeltaIndex: -1}
			}

			_, exist = delta.Balances[add]
			if exist == false {
				//Ensure we have balance for this block
				delta.Balances[add] = AddressBalance{Address: add, Balance: 0, LastDeltaIndex: bs.Balances[add].LastDeltaIndex}
			}

			tmp := delta.Balances[add]
			tmp.Balance += balance
			delta.Balances[add] = tmp
		}
		outs := tx.GetOutputs()
		for _, w := range outs {
			add := w.GetAddress().String()
			balance := int64(w.GetAmount())

			_, exist := bs.Balances[add]
			if exist == false {
				//Ensure we have top-level balance
				bs.Balances[add] = AddressBalance{Address: add, Balance: 0, LastDeltaIndex: -1}
			}

			_, exist = delta.Balances[add]
			if exist == false {
				//Ensure we have balance for this block
				delta.Balances[add] = AddressBalance{Address: add, Balance: 0, LastDeltaIndex: bs.Balances[add].LastDeltaIndex}
			}
			tmp := delta.Balances[add]
			tmp.Balance += balance
			delta.Balances[add] = tmp
		}
	}

	for k, v := range delta.Balances {
		tmp := bs.Balances[k]
		tmp.Balance += v.Balance
		tmp.LastDeltaIndex = index
		bs.Balances[k] = tmp
	}

	bs.Deltas = append(bs.Deltas, delta)

	return nil
}

type BalanceDelta struct {
	Balances map[string]AddressBalance
}

func (bs *BalanceDelta) Init() {
	if bs.Balances == nil {
		bs.Balances = map[string]AddressBalance{}
	}
}

type AddressBalance struct {
	Address        string
	Balance        int64
	LastDeltaIndex int //Last time this address' balance was changed
}
