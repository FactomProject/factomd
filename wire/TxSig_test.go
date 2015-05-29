// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"testing"	
	"bytes"
//	"fmt"
)

// Test TxSig read and write
func TestTxSig(t *testing.T) {

	txSig := new (TxSig)
	
	txSig.bitfield = []byte {1, 3}
	txSig.signatures = make([][]byte, 2)
	txSig.signatures[0] = []byte {1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 
								1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
								1, 2, 3, 4}
	txSig.signatures[1] = []byte {5, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 
								1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
								1, 2, 3, 4}
	
	var buf1 bytes.Buffer
	err:=writeSig(&buf1, 0, txSig)
	bytes1 := make ([]byte, len(buf1.Bytes()))
	copy(bytes1, buf1.Bytes())
	
	//fmt.Printf("txSig.signatures[1]:%v\n", txSig.signatures[1])	
	//fmt.Printf("buf1:%v\n", buf1.Bytes())	
	
	txSig2 := new (TxSig)
	readSig(&buf1, 0, 2, txSig2)
	
	//fmt.Printf("txSig2.signatures[1]:%v\n", txSig2.signatures[1])	
	
	var buf2 bytes.Buffer
	err=writeSig(&buf2, 0, txSig2)	
	bytes2 := make ([]byte, len(buf2.Bytes()))
	copy(bytes2, buf2.Bytes())	
	//fmt.Printf("buf2:%v\n:", buf2.Bytes())
	if err != nil{
		t.Errorf("signature did not match after unmarshalbinary:", err.Error())
	}
	
	if bytes.Compare(bytes1, bytes2) != 0 {
		t.Errorf("Invalid output")
	}	

	return
}

