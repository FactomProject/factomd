// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import (
    sc "github.com/FactomProject/simplecoin"
    "github.com/FactomProject/simplecoin/wallet"
    "time"
)

var _ = sc.Prt

var genesisBlock ISCBlock

func GetGenesisBlock() ISCBlock {
	if genesisBlock != nil {
		return genesisBlock
	}
	genesisBlock = NewSCBlock(1000000, uint32(time.Now().Unix()))  
    
    w := new(wallet.SCWallet)        
    t := w.CreateTransaction()
    
    for i:=0; i<10; i++ {
        h,_ := w.GenerateAddress([]byte("test "+string(i)),1,1)
        w.AddOutput(t,h,200000000000)   // 2000 factoids per address
    }
    
	flg, err := genesisBlock.AddCoinbase(t)
	if !flg || err != nil { 
        sc.Prtln("Flag: ",flg," Error: ",err)
    }
	return genesisBlock
}
