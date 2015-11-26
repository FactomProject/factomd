// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

type DBHead struct {
	KeyMR string
}

type RawData struct {
	Data string
}

type EBlockAddr struct {
	ChainID string
	KeyMR   string
}

type DBlock struct {
	Header struct {
		PrevBlockKeyMR string
		SequenceNumber uint32
		Timestamp      uint32
	}
	EntryBlockList []EBlockAddr
}

type EntryAddr struct {
	EntryHash string
	Timestamp uint32
}

type EBlock struct {
	Header struct {
		BlockSequenceNumber uint32
		ChainID             string
		PrevKeyMR           string
		Timestamp           uint32
	}
	EntryList []EntryAddr
}

type EntryStruct struct {
	ChainID string
	Content string
	ExtIDs  []string
}

type CHead struct {
	ChainHead string
}

type FactoidBalance struct {
	Response string
	Success  bool
}
