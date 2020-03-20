// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package constants

const (
	// MaxAckHeightDelta is the maximum number of blocks in the
	//	future we will set our HighestAckHeight too. This means
	// 	200 = max number of blocks to set the max height too on top
	//	of our current block height.
	MaxAckHeightDelta = 200
)

// Ack status levels
const (
	_ int = iota
	AckStatusInvalid
	AckStatusUnknown
	AckStatusNotConfirmed
	AckStatusACK
	AckStatus1Minute
	AckStatusDBlockConfirmed
)

// String forms of acks returned to users
const (
	AckStatusInvalidString         = "Invalid"
	AckStatusUnknownString         = "Unknown"
	AckStatusNotConfirmedString    = "NotConfirmed"
	AckStatusACKString             = "TransactionACK"
	AckStatus1MinuteString         = "1Minute"
	AckStatusDBlockConfirmedString = "DBlockConfirmed"
)

// AckStatusString will convert the input status int to a human readable string
func AckStatusString(status int) string {
	switch status {
	case AckStatusInvalid:
		return AckStatusInvalidString
	case AckStatusUnknown:
		return AckStatusUnknownString
	case AckStatusNotConfirmed:
		return AckStatusNotConfirmedString
	case AckStatusACK:
		return AckStatusACKString
	case AckStatus1Minute:
		return AckStatus1MinuteString
	case AckStatusDBlockConfirmed:
		return AckStatusDBlockConfirmedString
	}
	return "na"
}
