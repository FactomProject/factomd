// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This is a minimum wallet to be used to test the coin
// There isn't much in the way of interest in security 
// here, but rather provides a mechanism to create keys
// and sign transactions, etc.

package wallet

import (
    "fmt"
    "bytes"
    "encoding/binary"
    "github.com/FactomProject/simplecoin"
    "encoding/hex"
)

type IWalletEntry interface {
    simplecoin.IBlock
    SetRCD(simplecoin.IRCD)
    GetRCD() simplecoin.IRCD 
    AddKey(public, private []byte)
    GetName() ([]byte)  
    SetName([]byte)  
    GetAddress() (simplecoin.IHash, error)
    
}

type WalletEntry struct {
    IWalletEntry
    // 2 byte length not included here
    name    []byte          
    rcd     simplecoin.IRCD // Verification block for this IWalletEntry
    // 1 byte count of public keys
    public  [][]byte        // Set of public keys necessary towe sign the rcd
    // 1 byte count of private keys
    private [][]byte        // Set of private keys necessary to sign the rcd
}

var _ IWalletEntry = (*WalletEntry)(nil)

func (w1 WalletEntry)GetAddress() (simplecoin.IHash, error) {
    if w1.rcd == nil {
        return nil, fmt.Errorf("Should never happen. Missing the rcd block")
    }
    adr, err := w1.rcd.GetAddress()
    if err != nil {
        return nil, err
    }
    return adr, nil
}


func (w1 WalletEntry)GetDBHash() simplecoin.IHash {
    return simplecoin.Sha([]byte("WalletEntry")     )
}

func (w1 WalletEntry)GetNewInstance() simplecoin.IBlock {
    return new(WalletEntry)
}

func (w1 WalletEntry) IsEqual(w simplecoin.IBlock) bool {
    w2, ok := w.(*WalletEntry)
    if !ok { return false }
    
    for i, public := range w1.public {
        if bytes.Compare(w2.public[i],public) != 0 {
            return false
        }
    }
    return true 
}

func (w *WalletEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {
    
    len, data := binary.BigEndian.Uint16(data[0:2]), data[2:]
    n := make([]byte,len,len)   // build a place for the name
    copy(n,data[:len])          // copy it into that place
    data = data[len:]           // update data pointer
    w.name = n                  // Finally!  set the name
    
    if w.rcd == nil {
        w.rcd = simplecoin.CreateRCD(data)      // looks ahead, and creates the right RCD
    }   
    data,err := w.rcd.UnmarshalBinaryData(data)
    if err != nil { return nil, err }
    
    blen, data := data[0], data[1:]
    w.public = make([][]byte,len,len)
    for i:=0;i<int(blen);i++ {
        w.public[i] = make([]byte,simplecoin.ADDRESS_LENGTH,simplecoin.ADDRESS_LENGTH)
        copy(w.public[i],data[:simplecoin.ADDRESS_LENGTH])
        data = data[simplecoin.ADDRESS_LENGTH:]
    }
    
    blen, data = data[0], data[1:]
    w.private = make([][]byte,len,len)
    for i:=0;i<int(blen);i++ {
        w.private[i] = make([]byte,simplecoin.PRIVATE_LENGTH,simplecoin.PRIVATE_LENGTH)
        copy(w.private[i],data[:simplecoin.PRIVATE_LENGTH])
        data = data[simplecoin.PRIVATE_LENGTH:]
    }
    return data, nil
}

func (w WalletEntry) MarshalBinary() ([]byte, error) {
    var out bytes.Buffer
    
    binary.Write(&out, binary.BigEndian, uint16(len([]byte(w.name))))
    out.Write([]    byte(w.name))
    data,err := w.rcd.MarshalBinary()
    if err != nil {
        return nil, err
    }
    out.Write(data)
    out.WriteByte(byte(len(w.public)))
    for _,public := range w.public {
        out.Write(public)
    }
    out.WriteByte(byte(len(w.private)))
    for _,private := range w.private {
        out.Write(private)
    }
    return out.Bytes()  ,nil
}

func (w WalletEntry) MarshalText() (text []byte, err error) {
    var out bytes.Buffer

    out.WriteString("name:  ")
    out.Write   (w.name)
    out.WriteString("\n factoid address:")
    hash,err := w.rcd.GetAddress()
    out.WriteString(hash.String())
    out.WriteString("\n")
 
    out.WriteString("\n public:  ")
    for i,public := range w.public {
        simplecoin.WriteNumber16(&out, uint16(i))
        out.WriteString(" ")
        addr := hex.EncodeToString(public)
        out.WriteString(addr)
        out.WriteString("\n")
    }

    out.WriteString("\n private:  ")
    for i,private := range w.private {
        simplecoin.WriteNumber16(&out, uint16(i))
        out.WriteString(" ")
        addr := hex.EncodeToString(private)
        out.WriteString(addr)
        out.WriteString("\n")
    }
    
    return out.Bytes(), nil
}

func (w *WalletEntry) SetRCD(rcd simplecoin.IRCD) {
    w.rcd = rcd
}

func (w WalletEntry) GetRCD() simplecoin.IRCD  {
    return w.rcd
}


func (w *WalletEntry) AddKey(public, private []byte) {
    if len(public) != simplecoin.ADDRESS_LENGTH || 
       len(private) != simplecoin.PRIVATE_LENGTH {
        panic("Bad Keys presented to AddKey.  Should not happen.")
    }
    pu := make([]byte,simplecoin.ADDRESS_LENGTH,simplecoin.ADDRESS_LENGTH)
    pr := make([]byte,simplecoin.PRIVATE_LENGTH,simplecoin.PRIVATE_LENGTH)
    copy(pu,public)
    copy(pr,private)
    w.public = append(w.public,pu)
    w.private = append(w.private, pr)
    
    w.rcd = simplecoin.NewRCD_1(pu)
}

func (w *WalletEntry) SetName(name []byte) {
    w.name = name
}

