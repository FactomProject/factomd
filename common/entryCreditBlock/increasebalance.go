// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type IncreaseBalance struct {
	ECPubKey *primitives.ByteSlice32
	TXID     interfaces.IHash
	Index    uint64
	NumEC    uint64
}

var _ interfaces.Printable = (*IncreaseBalance)(nil)

var _ interfaces.BinaryMarshallable = (*IncreaseBalance)(nil)
var _ interfaces.ShortInterpretable = (*IncreaseBalance)(nil)
var _ interfaces.IECBlockEntry = (*IncreaseBalance)(nil)

func (e *IncreaseBalance) Init() {
	if e.ECPubKey == nil {
		e.ECPubKey = new(primitives.ByteSlice32)
	}
	if e.TXID == nil {
		e.TXID = primitives.NewZeroHash()
	}
}

func (a *IncreaseBalance) IsSameAs(b interfaces.IECBlockEntry) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	if a.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*IncreaseBalance)
	if ok == false {
		return false
	}

	if a.ECPubKey.IsSameAs(bb.ECPubKey) == false {
		return false
	}
	if a.TXID.IsSameAs(bb.TXID) == false {
		return false
	}
	if a.Index != bb.Index {
		return false
	}
	if a.NumEC != bb.NumEC {
		return false
	}

	return true
}

func (e *IncreaseBalance) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %-20s\n", "IncreaseBalance"))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "ECPubKey", e.ECPubKey[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "TXID", e.TXID.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Index", e.Index))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "NumEC", e.NumEC))

	return (string)(out.DeepCopyBytes())
}

func NewIncreaseBalance() *IncreaseBalance {
	r := new(IncreaseBalance)
	r.Init()
	return r
}

func (a *IncreaseBalance) GetEntryHash() interfaces.IHash {
	return nil
}

func (e *IncreaseBalance) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (e *IncreaseBalance) GetHash() interfaces.IHash {
	return e.Hash()
}

func (e *IncreaseBalance) GetSigHash() interfaces.IHash {
	return nil
}

func (b *IncreaseBalance) ECID() byte {
	return ECIDBalanceIncrease
}

func (b *IncreaseBalance) IsInterpretable() bool {
	return false
}

func (b *IncreaseBalance) Interpret() string {
	return ""
}

func (b *IncreaseBalance) MarshalBinary() ([]byte, error) {
	b.Init()
	buf := new(primitives.Buffer)

	buf.Write(b.ECPubKey[:])

	buf.Write(b.TXID.Bytes())

	primitives.EncodeVarInt(buf, b.Index)

	primitives.EncodeVarInt(buf, b.NumEC)

	return buf.DeepCopyBytes(), nil
}

func (b *IncreaseBalance) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling IncreaseBalance: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)
	hash := make([]byte, 32)

	_, err = buf.Read(hash)
	if err != nil {
		return
	}
	b.ECPubKey = new(primitives.ByteSlice32)
	copy(b.ECPubKey[:], hash)

	_, err = buf.Read(hash)
	if err != nil {
		return
	}
	if b.TXID == nil {
		b.TXID = primitives.NewZeroHash()
	}
	b.TXID.SetBytes(hash)

	tmp := make([]byte, 0)
	b.Index, tmp = primitives.DecodeVarInt(buf.DeepCopyBytes())

	b.NumEC, tmp = primitives.DecodeVarInt(tmp)

	newData = tmp
	return
}

func (b *IncreaseBalance) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

func (e *IncreaseBalance) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *IncreaseBalance) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *IncreaseBalance) GetTimestamp() interfaces.Timestamp {
	return nil
}
