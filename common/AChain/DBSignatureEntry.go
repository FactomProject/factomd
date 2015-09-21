package AChain

import ()

// DB Signature Entry -------------------------
type DBSignatureEntry struct {
	entryType            byte
	IdentityAdminChainID *Hash
	PubKey               PublicKey
	PrevDBSig            *Sig
}

var _ ABEntry = (*DBSignatureEntry)(nil)
var _ BinaryMarshallable = (*DBSignatureEntry)(nil)

// Create a new DB Signature Entry
func NewDBSignatureEntry(identityAdminChainID *Hash, sig Signature) (e *DBSignatureEntry) {
	e = new(DBSignatureEntry)
	e.entryType = TYPE_DB_SIGNATURE
	e.IdentityAdminChainID = identityAdminChainID
	e.PubKey = sig.Pub
	e.PrevDBSig = (*Sig)(sig.Sig)
	return
}

func (e *DBSignatureEntry) Type() byte {
	return e.entryType
}

func (e *DBSignatureEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	data, err = e.IdentityAdminChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	_, err = buf.Write(e.PubKey.Key[:])
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(e.PrevDBSig[:])
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *DBSignatureEntry) MarshalledSize() uint64 {
	var size uint64 = 0
	size += 1 // Type (byte)
	size += uint64(HASH_LENGTH)
	size += uint64(HASH_LENGTH)
	size += uint64(SIG_LENGTH)

	return size
}

func (e *DBSignatureEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	e.entryType, newData = newData[0], newData[1:]

	e.IdentityAdminChainID = new(Hash)
	newData, err = e.IdentityAdminChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	e.PubKey.Key = new([HASH_LENGTH]byte)
	copy(e.PubKey.Key[:], newData[:HASH_LENGTH])
	newData = newData[HASH_LENGTH:]

	e.PrevDBSig = new(Sig)
	copy(e.PrevDBSig[:], newData[:SIG_LENGTH])

	newData = newData[SIG_LENGTH:]

	return
}

func (e *DBSignatureEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *DBSignatureEntry) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *DBSignatureEntry) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *DBSignatureEntry) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (e *DBSignatureEntry) Spew() string {
	return Spew(e)
}

func (e *DBSignatureEntry) IsInterpretable() bool {
	return false
}

func (e *DBSignatureEntry) Interpret() string {
	return ""
}

func (e *DBSignatureEntry) Hash() *Hash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return Sha(bin)
}
