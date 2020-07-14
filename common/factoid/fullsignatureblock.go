package factoid

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type FullSignatureBlock struct {
	Signatures []interfaces.IFullSignature `json:"signatures"`
}

var _ interfaces.IFullSignatureBlock = (*FullSignatureBlock)(nil)

func (fsb *FullSignatureBlock) AddSignature(sig interfaces.IFullSignature) {
	fsb.Signatures = append(fsb.Signatures, sig)
}

func (fsb *FullSignatureBlock) GetSignature(index int) interfaces.IFullSignature {
	if index < 0 || index >= len(fsb.Signatures) {
		return nil
	}
	return fsb.Signatures[index]
}
func (fsb *FullSignatureBlock) GetSignatures() []interfaces.IFullSignature {
	return fsb.Signatures
}
func (fsb *FullSignatureBlock) IsSameAs(other interfaces.IFullSignatureBlock) bool {
	sigs := other.GetSignatures()
	if len(fsb.Signatures) != len(sigs) {
		return false
	}

	for i := range fsb.Signatures {
		if !fsb.Signatures[i].IsSameAs(sigs[i]) {
			return false
		}
	}
	return true
}

func (fsb *FullSignatureBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(fsb)
}

func (fsb *FullSignatureBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(fsb)
}

func (fsb FullSignatureBlock) MarshalBinary() ([]byte, error) {
	sigs := fsb.GetSignatures()
	buf := primitives.NewBuffer(nil)
	buf.PushVarInt(uint64(len(sigs)))
	for _, sig := range sigs {
		err := buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}
	return buf.DeepCopyBytes(), nil
}

func (fsb *FullSignatureBlock) UnmarshalBinary(data []byte) error {
	if _, err := fsb.UnmarshalBinaryData(data); err != nil {
		return err
	}
	return nil
}

func (fsb *FullSignatureBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	length, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}

	fsb.Signatures = make([]interfaces.IFullSignature, length)
	for i := uint64(0); i < length; i++ {
		fsb.Signatures[i] = new(primitives.Signature)
		if err := buf.PopBinaryMarshallable(fsb.Signatures[i]); err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (fsb *FullSignatureBlock) String() string {
	var out primitives.Buffer

	out.WriteString("Signature Block: \n")
	for _, sig := range fsb.Signatures {
		out.WriteString(" signature: ")
		if txt, err := sig.CustomMarshalText(); err != nil {
			out.WriteString("<error> ")
			out.WriteString(err.Error())
		} else {
			out.Write(txt)
		}
		out.WriteString("\n ")
	}

	return out.String()
}

func NewFullSignatureBlock() *FullSignatureBlock {
	fsb := new(FullSignatureBlock)
	return fsb
}
