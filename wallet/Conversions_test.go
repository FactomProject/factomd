package wallet

import (
	"fmt"
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

func TestHumanReadiblyPrivateKeyToPrivateKey(t *testing.T) {
	priv, err := HumanReadableFactoidPrivateKeyToPrivateKey("Fs37iVGnZ7jShPudsXuB98qURxk35eLrmh9cgPuPpTXHAJEBkUTh")
	if err != nil {
		t.Error(err)
	}
	if fmt.Sprintf("%x", priv) != "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973" {
		t.Errorf("Incorrect private key returned")
	}
}
