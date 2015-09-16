package wallet

import (
	"fmt"
	"github.com/FactomProject/go-bip39"
	"testing"
)

func TestMnemonicStringToPrivateKey(t *testing.T) {
	priv, err := MnemonicStringToPrivateKey("salute umbrella proud setup delay ginger practice split toss jewel tuition stool")
	if err != nil {
		t.Error(err)
	}
	if fmt.Sprintf("%x", priv) != "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973" {
		t.Errorf("Incorrect private key returned")
	}
}

func TestHumanReadablePrivateKeyToPrivateKey(t *testing.T) {
	priv, err := HumanReadableFactoidPrivateKeyToPrivateKey("Fs37iVGnZ7jShPudsXuB98qURxk35eLrmh9cgPuPpTXHAJEBkUTh")
	if err != nil {
		t.Error(err)
	}
	if fmt.Sprintf("%x", priv) != "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973" {
		t.Errorf("Incorrect private key returned")
	}
}

func TestMnemonicValidation(t *testing.T) {
	properMnemonic := "salute umbrella proud setup delay ginger practice split toss jewel tuition stool"
	improperMnemonics := []string{
		"salute umbrella proud setup delay ginger practice split toss jewel tuition story",
		"salute umbrella proud setup delay ginger practice split toss jewel tuition",
		"xalute umbrella proud setup delay ginger practice split toss jewel tuition stool",
	}

	priv, err := MnemonicStringToPrivateKey(improperMnemonics[0])
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
	if fmt.Sprintf("%x", priv) == "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973" {
		t.Errorf("We received the same private key even though the mnemonic is wrong")
	}

	_, err = bip39.MnemonicToByteArray(properMnemonic)
	if err != nil {
		t.Errorf("Error during mnemonic conversion - %v", err)
	}

	for _, improperMnemonic := range improperMnemonics {
		_, err := MnemonicStringToPrivateKey(improperMnemonic)
		if err == nil {
			t.Errorf("Error is nil when it shouldn't be")
		}
		_, err = bip39.MnemonicToByteArray(improperMnemonic)
		if err == nil {
			t.Errorf("Received no error for invalid mnemonic %v", improperMnemonic)
		}
	}
}
