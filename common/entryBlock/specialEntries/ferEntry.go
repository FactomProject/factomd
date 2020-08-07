// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package specialEntries

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

// FEREntry stands for 'Factoid Exchange Rate' (FER) entry.
type FEREntry struct {
	Version string `json:"version"` // This appears unused

	// The directory block height the FER request will be cancelled if not performed. Must be >= current directory block height
	// and <= current directory block height + 12
	ExpirationHeight uint32 `json:"expiration_height"`

	// The directory block height set by the state machine when the entry is processed. Used only for validity checking the expiration height
	ResidentHeight uint32 `json:"resident_height"`

	// The directory block height the FER change should occur. Must be >= expiration height - 6
	TargetActivationHeight uint32 `json:"target_activation_height"`

	Priority    uint32 `json:"priority"`     // Higher priorities take precedence over lower priorities
	TargetPrice uint64 `json:"target_price"` // The actual FER in Factoshis per EC (Entry Credit)
}

var _ interfaces.Printable = (*FEREntry)(nil)
var _ interfaces.BinaryMarshallable = (*FEREntry)(nil)
var _ interfaces.IFEREntry = (*FEREntry)(nil)

// GetVersion returns the version
func (fer *FEREntry) GetVersion() string {
	return fer.Version
}

// SetVersion sets the version to the input value
func (fer *FEREntry) SetVersion(passedVersion string) interfaces.IFEREntry {
	fer.Version = passedVersion
	return fer
}

// GetExpirationHeight returns the expiration height
func (fer *FEREntry) GetExpirationHeight() uint32 {
	return fer.ExpirationHeight
}

// SetExpirationHeight sets the expiration height to the input value
func (fer *FEREntry) SetExpirationHeight(passedExpirationHeight uint32) interfaces.IFEREntry {
	fer.ExpirationHeight = passedExpirationHeight
	return fer
}

// GetResidentHeight returns the resident height
func (fer *FEREntry) GetResidentHeight() uint32 {
	return fer.ResidentHeight
}

// SetResidentHeight sets the resident height to the input value
func (fer *FEREntry) SetResidentHeight(passedResidentHeight uint32) interfaces.IFEREntry {
	fer.ResidentHeight = passedResidentHeight
	return fer
}

// GetTargetActivationHeight returns the target activation height
func (fer *FEREntry) GetTargetActivationHeight() uint32 {
	return fer.TargetActivationHeight
}

// SetTargetActivationHeight sets the target activation height to the input value
func (fer *FEREntry) SetTargetActivationHeight(passedTargetActivationHeight uint32) interfaces.IFEREntry {
	fer.TargetActivationHeight = passedTargetActivationHeight
	return fer
}

// GetPriority returns the priority
func (fer *FEREntry) GetPriority() uint32 {
	return fer.Priority
}

// SetPriority sets the priority to the input value
func (fer *FEREntry) SetPriority(passedPriority uint32) interfaces.IFEREntry {
	fer.Priority = passedPriority
	return fer
}

// GetTargetPrice returns the target price
func (fer *FEREntry) GetTargetPrice() uint64 {
	return fer.TargetPrice
}

// SetTargetPrice sets the target price to the input value
func (fer *FEREntry) SetTargetPrice(passedTargetPrice uint64) interfaces.IFEREntry {
	fer.TargetPrice = passedTargetPrice
	return fer
}

// JSONByte returns the json encoded byte array
func (fer *FEREntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(fer)
}

// JSONString returns the json encoded string
func (fer *FEREntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(fer)
}

// String returns this object as a string
func (fer *FEREntry) String() string {
	str, _ := fer.JSONString()
	return str
}

// UnmarshalBinaryData unmarshals the input data into this object
func (fer *FEREntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	return nil, json.Unmarshal(data, fer)
}

// UnmarshalBinary unmarshals the input data into this object
func (fer *FEREntry) UnmarshalBinary(data []byte) (err error) {
	_, err = fer.UnmarshalBinaryData(data)
	return
}

// MarshalBinary marshals this object
func (fer *FEREntry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FEREntry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return json.Marshal(fer)
}
