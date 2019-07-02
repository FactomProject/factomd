// +build all

package factoid_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
)

/*

Please enter the 12 Koinify words: salute umbrella proud setup delay ginger practice split toss jewel tuition stool
seed derived from words: 2baa21c5b5cae271225d8b3a0fd3833a384cb0c989d785dc2424b14d6a6d5c7bd7e9c6ed4bbe458c006cbc196566e414d845aeab7983de710d634fc371f0b640
BIP32 root key: 0488ade4000000000000000000dafb93929fd40b7740d9e99848c988d16d6571c992c098587b1fc8849ab54aa100d6b0bb285dc87d07549e08d85e39ec6a81f78aaedef38304ec7779fa7cb5665f
BIP32 root of Factoid chain key: 0488ade401251c4d7080000007b4972ca4cdea08295da0e08f7cdb5642e2cd976e349f4115b775e343424248ef00ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973
Last 32 bytes, ed25519 private key: ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973
Private key with prefix:   6478ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973
Private key hash: fabdae1072e2a27bf250e647ef01dc390e2b15338d198f7039fb32bcf80aead5
Private key with checksum: 6478ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973fabdae10
Human readable private key: Fs37iVGnZ7jShPudsXuB98qURxk35eLrmh9cgPuPpTXHAJEBkUTh
ed25519 Factoid public key: 8bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4
test sig: 54d481170c29dbd8119b0ba234faafa0ecc9a233c3f36007310336a8ef806f77b84cdff28d4b1b254590bf77878d2c90895a0280e27d27f39aca3932392a5b08
signature good
data encoded in OP_RETURN is: 464143544f4d30308bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4
Private key from the Koinify words: Fs37iVGnZ7jShPudsXuB98qURxk35eLrmh9cgPuPpTXHAJEBkUTh
SHA256 hash of pubkey: 673e4e11ea4d647f60a1ea36f7f3102616172a82e437d797bade1730a47bd133
first 4 bytes of pubkey hash: 673e4e11
pubkey with checksum:  8bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4673e4e11
Corresponding to public key: 8bee2930-cbe4772a-e5454c48-01d4ef36-6276f6e4-cc65bac1-8be03607-c00288c4-673e4e11

*/

func TestPublicToAddress(t *testing.T) {
	add, err := PublicKeyStringToFactoidAddressString("8bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4")
	if err != nil {
		t.Error(err)
	}
	if add != "FA3Y1tBWnFpyoZUPr9ZH51R1gSC8r5x5kqvkXL3wy4uRvzFnuWLB" {
		t.Error("Wrong address returned")
	}
}

func TestCombined(t *testing.T) {
	priv, pub, add, err := HumanReadablePrivateKeyStringToEverythingString("Fs37iVGnZ7jShPudsXuB98qURxk35eLrmh9cgPuPpTXHAJEBkUTh")
	if err != nil {
		t.Error(err)
	}
	if priv != "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973" {
		t.Error("Wrong private key returned")
	}
	if pub != "8bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4" {
		t.Error("Wrong public key returned")
	}
	if add != "FA3Y1tBWnFpyoZUPr9ZH51R1gSC8r5x5kqvkXL3wy4uRvzFnuWLB" {
		t.Error("Wrong address returned")
	}
}

func TestPublicKeyStringToFactoidAddress(t *testing.T) {
	add, err := PublicKeyStringToFactoidAddress("8bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4")
	if err != nil {
		t.Errorf("%v", err)
	}
	if primitives.ConvertFctAddressToUserStr(add) != "FA3Y1tBWnFpyoZUPr9ZH51R1gSC8r5x5kqvkXL3wy4uRvzFnuWLB" {
		t.Errorf("Wrong address returned - %v", primitives.ConvertFctAddressToUserStr(add))
	}
}

func TestPublicKeyStringToECAddress(t *testing.T) {
	add, err := PublicKeyStringToECAddress("8bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4")
	if err != nil {
		t.Errorf("%v", err)
	}
	if primitives.ConvertECAddressToUserStr(add) != "EC2pn759osDzgF4BH5GsXw4NkgTMTrY6jQ7nhhvqkMoBmzjB4Soq" {
		t.Errorf("Wrong address returned - %v", primitives.ConvertECAddressToUserStr(add))
	}
}

func TestECAddress(t *testing.T) {
	add, err := PublicKeyStringToECAddressString("8bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4")
	if err != nil {
		t.Error(err)
	}
	if add != "EC2pn759osDzgF4BH5GsXw4NkgTMTrY6jQ7nhhvqkMoBmzjB4Soq" {
		t.Errorf("Wrong address returned - %v", add)
	}
}
