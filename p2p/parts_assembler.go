// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"time"
)

// maximum time we wait for a partial message to arrive, old entries are cleaned up only when new part arrives
const MaxTimeWaitingForReassembly time.Duration = time.Second * 60 * 10

type PartialMessage struct {
	parts                  []*Parcel // array of message parts we've received so far
	firstPartReceived      time.Time // a timestamp indicating when the first part was received
	mostRecentPartReceived time.Time // a timestamp indicating when the mostRecent part was received
}

// PartsAssembler is responsible for assembling message parts into full messages
type PartsAssembler struct {
	messages map[string]*PartialMessage // a map of app hashes to partial messages
}

// Initializes the assembler
func (assembler *PartsAssembler) Init() *PartsAssembler {
	assembler.messages = make(map[string]*PartialMessage)
	return assembler
}

// Handles a single message part, returns either a fully assembled message or nil
func (assembler *PartsAssembler) handlePart(parcel Parcel) *Parcel {
	debug("PartsAssembler", "Handling message part %s %d/%d", parcel.Header.AppHash, parcel.Header.PartNo, parcel.Header.PartsTotal)
	partial, exists := assembler.messages[parcel.Header.AppHash]

	valid, err := validateParcelPart(parcel, partial)
	if !valid {
		significant("PartsAssembler", "Detected invalid parcel: %s, dropping", err)
		return nil
	}

	if !exists {
		partial = createNewPartialMessage(parcel)
		assembler.messages[parcel.Header.AppHash] = partial
	}

	partial.parts[parcel.Header.PartNo] = &parcel
	partial.mostRecentPartReceived = time.Now()

	// get an assembled parcel or nil if not yet ready
	fullParcel := tryReassemblingMessage(partial)
	if fullParcel != nil {
		delete(assembler.messages, parcel.Header.AppHash)
		debug("PartsAssembler", "Fully assembled %s", parcel.Header.AppHash)
	}

	// go through all partial messages and removes the old ones
	assembler.cleanupOldPartialMessages()

	return fullParcel
}

// checks if part is valid for assembler to process
func validateParcelPart(parcel Parcel, partial *PartialMessage) (bool, string) {
	if parcel.Header.PartsTotal <= 0 {
		return false, "PartsTotal less or equal 0"
	}

	if parcel.Header.PartNo < 0 {
		return false, "PartNo less than 0"
	}

	if parcel.Header.PartNo >= parcel.Header.PartsTotal {
		return false, "PartNo outside of PartsTotal range"
	}

	if partial != nil && parcel.Header.PartsTotal != uint16(len(partial.parts)) {
		return false, "PartsTotal does not match allocated array of parts"
	}

	return true, "" // valid
}

// Checks existing partial messages and if there is anything older than MaxTimeWaitingForReassembly,
// drops the partial message
func (assembler *PartsAssembler) cleanupOldPartialMessages() {
	for appHash, partial := range assembler.messages {
		timeWaiting := time.Since(partial.mostRecentPartReceived)
		timeSinceFirst := time.Since(partial.firstPartReceived)
		if timeWaiting > MaxTimeWaitingForReassembly {
			delete(assembler.messages, appHash)
			note("PartsAssembler", "Dropping message %d after %s secs, time since first part: %s secs",
				appHash, timeWaiting/time.Second, timeSinceFirst/time.Second)
		}
	}
}

// Creates a new PartialMessage from a given parcel
func createNewPartialMessage(parcel Parcel) *PartialMessage {
	partial := new(PartialMessage)
	partial.parts = make([]*Parcel, parcel.Header.PartsTotal)
	partial.firstPartReceived = time.Now()
	return partial
}

// Tries reassembling a full Message from existing MessageParts, returns nil if
// we don't have all the necessary parts yet
func tryReassemblingMessage(partial *PartialMessage) *Parcel {
	for _, part := range partial.parts {
		if part == nil {
			return nil
		}
	}

	return ReassembleParcel(partial.parts)
}
