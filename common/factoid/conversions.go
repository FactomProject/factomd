package factoid

import (
	"encoding/hex"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

/******************************************************************************/
/****************************To addresses**************************************/
/******************************************************************************/

// PublicKeyStringToECAddressString converts a hexidecimal public key string to a base58 EC address string
func PublicKeyStringToECAddressString(public string) (string, error) {
	pubHex, err := hex.DecodeString(public)
	if err != nil {
		return "", err
	}

	add, err := PublicKeyToECAddress(pubHex)
	if err != nil {
		return "", err
	}

	return primitives.ConvertECAddressToUserStr(add), nil
}

// PublicKeyStringToECAddress converts a hexidecimal public key string to an EC address
func PublicKeyStringToECAddress(public string) (interfaces.IAddress, error) {
	pubHex, err := hex.DecodeString(public)
	if err != nil {
		return nil, err
	}
	return PublicKeyToECAddress(pubHex)
}

// PublicKeyToECAddress creates a new address from the public key
func PublicKeyToECAddress(public []byte) (interfaces.IAddress, error) {
	return NewAddress(public), nil
}

// PublicKeyStringToFactoidAddressString converts the input hexidecimal public key string to a base58 factoid address string
func PublicKeyStringToFactoidAddressString(public string) (string, error) {
	pubHex, err := hex.DecodeString(public)
	if err != nil {
		return "", err
	}
	add, err := PublicKeyToFactoidAddress(pubHex)
	if err != nil {
		return "", err
	}

	return primitives.ConvertFctAddressToUserStr(add), nil
}

// PublicKeyToFactoidAddress converts a public key to new factoid address
func PublicKeyToFactoidAddress(public []byte) (interfaces.IAddress, error) {
	rcd := NewRCD_1(public)
	add, err := rcd.GetAddress()
	if err != nil {
		return nil, err
	}
	return add, nil
}

// PublicKeyStringToFactoidAddress converts a hexidecimal public key string to a factoid address
func PublicKeyStringToFactoidAddress(public string) (interfaces.IAddress, error) {
	pubHex, err := hex.DecodeString(public)
	if err != nil {
		return nil, err
	}
	rcd := NewRCD_1(pubHex)
	add, err := rcd.GetAddress()
	if err != nil {
		return nil, err
	}
	return add, nil
}

// PublicKeyStringToFactoidRCDAddress converts a hexidecimal public key string to a factoid RCD address
func PublicKeyStringToFactoidRCDAddress(public string) (interfaces.IRCD, error) {
	pubHex, err := hex.DecodeString(public)
	if err != nil {
		return nil, err
	}
	rcd := NewRCD_1(pubHex)
	return rcd, nil
}

/******************************************************************************/
/******************************Combined****************************************/
/******************************************************************************/

// HumanReadablePrivateKeyStringToEverythingString converts a human readable base58 private key string to a hexidecimal public key
// string and a base58 factoid address string
func HumanReadablePrivateKeyStringToEverythingString(private string) (string, string, string, error) {
	priv, err := primitives.HumanReadableFactoidPrivateKeyToPrivateKeyString(private)
	if err != nil {
		return "", "", "", err
	}
	return PrivateKeyStringToEverythingString(priv)
}

// PrivateKeyStringToEverythingString converts a hexidecimal private key string to a hexidecimal public key string and
// a base58 factoid address string
func PrivateKeyStringToEverythingString(private string) (string, string, string, error) {
	pub, err := primitives.PrivateKeyStringToPublicKeyString(private)
	if err != nil {
		return "", "", "", err
	}
	add, err := PublicKeyStringToFactoidAddressString(pub)
	if err != nil {
		return "", "", "", err
	}
	return private, pub, add, nil
}
