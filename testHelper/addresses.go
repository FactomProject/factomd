package testHelper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/factoid/wallet"
	"github.com/FactomProject/factomd/common/interfaces"
)

func NewFactoidAddressStrings(n uint64) (string, string, string) {
	//ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973
	//000000000000000000000000000000000000000000000000XXXXXXXXXXXXXXXX
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, n); err != nil {
		panic(err)
	}

	priv := fmt.Sprintf("000000000000000000000000000000000000000000000000%x", buf.Bytes())
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
