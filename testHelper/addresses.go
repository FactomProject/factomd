package testHelper

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/factoid/wallet"
	"github.com/FactomProject/factomd/common/interfaces"
)

func NewPrivKeyString(n uint64) string {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, n); err != nil {
		panic(err)
	}

	priv := fmt.Sprintf("000000000000000000000000000000000000000000000000%x", buf.Bytes())
	return priv
}

func NewPrivKey(n uint64) []byte {
	priv := NewPrivKeyString(n)
	p, err := hex.DecodeString(priv)
	if err != nil {
		panic(err)
	}
	return p
}

func NewFactoidAddressStrings(n uint64) (string, string, string) {
	//ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973
	//000000000000000000000000000000000000000000000000XXXXXXXXXXXXXXXX
	priv := NewPrivKeyString(n)
	privKey, pubKey, add, err := wallet.PrivateKeyStringToEverythingString(priv)
	if err != nil {
		panic(err)
	}
	return privKey, pubKey, add
}

func NewFactoidAddress(n uint64) interfaces.IAddress {
	_, pub, _ := NewFactoidAddressStrings(n)
	add, err := wallet.PublicKeyStringToFactoidAddress(pub)
	if err != nil {
		panic(err)
	}
	return add
}

func NewFactoidRCDAddressString(n uint64) string {
	add := NewFactoidAddress(n)
	return add.String()
}

func NewFactoidRCDAddress(n uint64) interfaces.IRCD {
	_, pub, _ := NewFactoidAddressStrings(n)
	add, err := wallet.PublicKeyStringToFactoidRCDAddress(pub)
	if err != nil {
		panic(err)
	}
	return add
}

func NewECAddress(n uint64) interfaces.IAddress {
	_, pub, _ := NewFactoidAddressStrings(n)
	add, err := wallet.PublicKeyStringToFactoidAddress(pub)
	if err != nil {
		panic(err)
	}
	return add
}

func PrivateKeyToEDPub(priv []byte) []byte {
	priv2 := new([ed25519.PrivateKeySize]byte)
	copy(priv2[:], priv)
	pub := ed25519.GetPublicKey(priv2)
	return pub[:]
}
