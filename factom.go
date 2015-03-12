// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Glue code between BTCD code & Factom.

package main

import (
	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/wire"
	"github.com/FactomProject/btcutil"
)

// func (pl *ProcessList) AddFtmTxToProcessList(msg wire.Message, msgHash *wire.ShaHash) error {
func fakehook1(msg wire.Message, msgHash *wire.ShaHash) error {
	return nil
}

func factom_PL_hook(tx *btcutil.Tx, label string) error {
	util.Trace("label= " + label)

	_ = fakehook1(tx.MsgTx(), tx.Sha())

	return nil
}
