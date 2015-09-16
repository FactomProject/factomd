package wallet

import (
	"fmt"
	"github.com/FactomProject/go-bip32"
	"github.com/FactomProject/go-bip39"
	"github.com/btcsuitereleases/btcutil/base58"
	"strings"
)

func MnemonicStringToPrivateKey(mnemonic string) ([]byte, error) {
	mnemonic = strings.ToLower(strings.TrimSpace(mnemonic))
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	child, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 7)
	if err != nil {
		return nil, err
	}

	return child.Key, nil
}

func HumanReadableFactoidPrivateKeyToPrivateKey(human string) ([]byte, error) {
	human = strings.TrimSpace(human)
	base, version, err := base58.CheckDecode(human)
	if err != nil {
		return nil, err
	}

	if version != 0x64 || base[0] != 0x78 {
		return nil, fmt.Errorf("Invalid prefix")
	}

	return base[1:], nil
}

func HumanReadableECPrivateKeyToPrivateKey(human string) ([]byte, error) {
	human = strings.TrimSpace(human)
	base, version, err := base58.CheckDecode(human)
	if err != nil {
		return nil, err
	}

	if version != 0x5d || base[0] != 0xb6 {
		return nil, fmt.Errorf("Invalid prefix")
	}

	return base[1:], nil
}
