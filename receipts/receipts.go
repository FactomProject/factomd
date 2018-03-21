// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package receipts

import (
	"encoding/json"
	"fmt"

	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type Receipt struct {
	Entry                  *JSON                    `json:"entry,omitempty"`
	MerkleBranch           []*primitives.MerkleNode `json:"merklebranch,omitempty"`
	EntryBlockKeyMR        *primitives.Hash         `json:"entryblockkeymr,omitempty"`
	DirectoryBlockKeyMR    *primitives.Hash         `json:"directoryblockkeymr,omitempty"`
	BitcoinTransactionHash *primitives.Hash         `json:"bitcointransactionhash,omitempty"`
	BitcoinBlockHash       *primitives.Hash         `json:"bitcoinblockhash,omitempty"`
}

func (e *Receipt) TrimReceipt() {
	if e == nil {
		return
	}
	entry, _ := primitives.NewShaHashFromStr(e.Entry.EntryHash)
	for i := range e.MerkleBranch {
		if entry.IsSameAs(e.MerkleBranch[i].Left) {
			e.MerkleBranch[i].Left = nil
		} else {
			if entry.IsSameAs(e.MerkleBranch[i].Right) {
				e.MerkleBranch[i].Right = nil
			}
		}
		entry = e.MerkleBranch[i].Top
		e.MerkleBranch[i].Top = nil
	}
}

func (e *Receipt) Validate() error {
	if e == nil {
		return fmt.Errorf("No receipt provided")
	}
	if e.Entry == nil {
		return fmt.Errorf("Receipt has no entry")
	}
	if e.MerkleBranch == nil {
		return fmt.Errorf("Receipt has no MerkleBranch")
	}
	if e.EntryBlockKeyMR == nil {
		return fmt.Errorf("Receipt has no EntryBlockKeyMR")
	}
	if e.DirectoryBlockKeyMR == nil {
		return fmt.Errorf("Receipt has no DirectoryBlockKeyMR")
	}
	entryHash, err := primitives.NewShaHashFromStr(e.Entry.EntryHash)
	//TODO: validate entry hashes into EntryHash

	if err != nil {
		return err
	}
	var left interfaces.IHash
	var right interfaces.IHash
	var currentEntry interfaces.IHash
	currentEntry = entryHash
	eBlockFound := false
	dBlockFound := false
	for i, node := range e.MerkleBranch {
		if node.Left == nil {
			if node.Right == nil {
				return fmt.Errorf("Node %v/%v has two nil sides", i, len(e.MerkleBranch))
			}
			left = currentEntry
			right = node.Right
		} else {
			left = node.Left
			if node.Right == nil {
				right = currentEntry
			} else {
				right = node.Right
			}
		}
		if left.IsSameAs(currentEntry) == false && left.IsSameAs(currentEntry) {
			return fmt.Errorf("Entry %v not found in node %v/%v", currentEntry, i, len(e.MerkleBranch))
		}
		top := primitives.HashMerkleBranches(left, right)
		if node.Top != nil {
			if top.IsSameAs(node.Top) == false {
				return fmt.Errorf("Derived top %v is not the same as saved top in node %v/%v", top, i, len(e.MerkleBranch))
			}
		}
		if top.IsSameAs(e.EntryBlockKeyMR) == true {
			eBlockFound = true
		}
		if top.IsSameAs(e.DirectoryBlockKeyMR) == true {
			dBlockFound = true
		}
		currentEntry = top
	}

	if eBlockFound == false {
		return fmt.Errorf("EntryBlockKeyMR not found in branch")
	}

	if dBlockFound == false {
		return fmt.Errorf("DirectoryBlockKeyMR not found in branch")
	}

	return nil
}

func (e *Receipt) IsSameAs(r *Receipt) bool {
	if e.Entry == nil {
		if r.Entry != nil {
			return false
		}
	} else {
		if e.Entry.IsSameAs(r.Entry) == false {
			return false
		}
	}

	if e.MerkleBranch == nil {
		if r.MerkleBranch != nil {
			return false
		}
	} else {
		if len(e.MerkleBranch) != len(r.MerkleBranch) {
			return false
		}
		for i := range e.MerkleBranch {
			if e.MerkleBranch[i].Left == nil {
				if r.MerkleBranch[i].Left != nil {
					return false
				}
			} else {
				if e.MerkleBranch[i].Left.IsSameAs(r.MerkleBranch[i].Left) == false {
					return false
				}
			}
			if e.MerkleBranch[i].Right == nil {
				if r.MerkleBranch[i].Right != nil {
					return false
				}
			} else {
				if e.MerkleBranch[i].Right.IsSameAs(r.MerkleBranch[i].Right) == false {
					return false
				}
			}
			if e.MerkleBranch[i].Top == nil {
				if r.MerkleBranch[i].Top != nil {
					return false
				}
			} else {
				if e.MerkleBranch[i].Top.IsSameAs(r.MerkleBranch[i].Top) == false {
					return false
				}
			}
		}
	}

	if e.EntryBlockKeyMR == nil {
		if r.EntryBlockKeyMR != nil {
			return false
		}
	} else {
		if e.EntryBlockKeyMR.IsSameAs(r.EntryBlockKeyMR) == false {
			return false
		}
	}

	if e.DirectoryBlockKeyMR == nil {
		if r.DirectoryBlockKeyMR != nil {
			return false
		}
	} else {
		if e.DirectoryBlockKeyMR.IsSameAs(r.DirectoryBlockKeyMR) == false {
			return false
		}
	}

	if e.BitcoinTransactionHash == nil {
		if r.BitcoinTransactionHash != nil {
			return false
		}
	} else {
		if e.BitcoinTransactionHash.IsSameAs(r.BitcoinTransactionHash) == false {
			return false
		}
	}

	if e.BitcoinBlockHash == nil {
		if r.BitcoinBlockHash != nil {
			return false
		}
	} else {
		if e.BitcoinBlockHash.IsSameAs(r.BitcoinBlockHash) == false {
			return false
		}
	}

	return true
}

func (e *Receipt) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Receipt) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Receipt) CustomMarshalString() string {
	str, _ := e.JSONString()
	return str
}

func (e *Receipt) DecodeString(str string) error {
	jsonByte := []byte(str)
	err := json.Unmarshal(jsonByte, e)
	if err != nil {
		return err
	}
	return nil
}

func DecodeReceiptString(str string) (*Receipt, error) {
	receipt := new(Receipt)
	err := json.Unmarshal([]byte(str), &receipt)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

type JSON struct {
	Raw       string `json:"raw,omitempty"`
	EntryHash string `json:"entryhash,omitempty"`
	Json      string `json:"json,omitempty"`
}

func (e *JSON) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *JSON) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *JSON) String() string {
	str, _ := e.JSONString()
	return str
}

func (e *JSON) IsSameAs(r *JSON) bool {
	if r == nil {
		return false
	}
	if e.Raw != r.Raw {
		return false
	}
	if e.EntryHash != r.EntryHash {
		return false
	}
	if e.Json != r.Json {
		return false
	}
	return true
}

func CreateFullReceipt(dbo interfaces.DBOverlaySimple, entryID interfaces.IHash) (*Receipt, error) {
	return CreateReceipt(dbo, entryID)
}

func CreateMinimalReceipt(dbo interfaces.DBOverlaySimple, entryID interfaces.IHash) (*Receipt, error) {
	receipt, err := CreateReceipt(dbo, entryID)
	if err != nil {
		return nil, err
	}

	receipt.TrimReceipt()

	return receipt, nil
}

func CreateReceipt(dbo interfaces.DBOverlaySimple, entryID interfaces.IHash) (*Receipt, error) {
	receipt := new(Receipt)
	receipt.Entry = new(JSON)
	receipt.Entry.EntryHash = entryID.String()

	//EBlock

	hash, err := dbo.FetchIncludedIn(entryID)
	if err != nil {
		return nil, err
	}

	if hash == nil {
		return nil, fmt.Errorf("Block containing entry not found")
	}

	eBlock, err := dbo.FetchEBlock(hash)
	if err != nil {
		return nil, err
	}

	if eBlock == nil {
		return nil, fmt.Errorf("EBlock not found")
	}

	hash = eBlock.DatabasePrimaryIndex()
	receipt.EntryBlockKeyMR = hash.(*primitives.Hash)

	entries := eBlock.GetEntryHashes()
	//fmt.Printf("eBlock entries - %v\n\n", entries)
	branch := primitives.BuildMerkleBranchForEntryHash(entries, entryID, true)
	blockNode := new(primitives.MerkleNode)
	left, err := eBlock.HeaderHash()
	if err != nil {
		return nil, err
	}
	blockNode.Left = left.(*primitives.Hash)
	blockNode.Right = eBlock.BodyKeyMR().(*primitives.Hash)
	blockNode.Top = hash.(*primitives.Hash)
	//fmt.Printf("eBlock blockNode - %v\n\n", blockNode)
	branch = append(branch, blockNode)
	receipt.MerkleBranch = append(receipt.MerkleBranch, branch...)

	//str, _ := eBlock.JSONString()
	//fmt.Printf("eBlock - %v\n\n", str)

	//DBlock

	hash, err = dbo.FetchIncludedIn(hash)
	if err != nil {
		return nil, err
	}

	if hash == nil {
		return nil, fmt.Errorf("Block containing EBlock not found")
	}

	dBlock, err := dbo.FetchDBlock(hash)
	if err != nil {
		return nil, err
	}

	if dBlock == nil {
		return nil, fmt.Errorf("DBlock not found")
	}

	//str, _ = dBlock.JSONString()
	//fmt.Printf("dBlock - %v\n\n", str)

	entries = dBlock.GetEntryHashesForBranch()
	//fmt.Printf("dBlock entries - %v\n\n", entries)

	//merkleTree := primitives.BuildMerkleTreeStore(entries)
	//fmt.Printf("dBlock merkleTree - %v\n\n", merkleTree)

	branch = primitives.BuildMerkleBranchForEntryHash(entries, receipt.EntryBlockKeyMR, true)
	blockNode = new(primitives.MerkleNode)
	left, err = dBlock.HeaderHash()
	if err != nil {
		return nil, err
	}
	blockNode.Left = left.(*primitives.Hash)
	blockNode.Right = dBlock.BodyKeyMR().(*primitives.Hash)
	blockNode.Top = hash.(*primitives.Hash)
	//fmt.Printf("dBlock blockNode - %v\n\n", blockNode)
	branch = append(branch, blockNode)
	receipt.MerkleBranch = append(receipt.MerkleBranch, branch...)

	//DirBlockInfo

	hash = dBlock.DatabasePrimaryIndex()
	receipt.DirectoryBlockKeyMR = hash.(*primitives.Hash)

	dirBlockInfo, err := dbo.FetchDirBlockInfoByKeyMR(hash)
	if err != nil {
		return nil, err
	}

	if dirBlockInfo != nil {
		dbi := dirBlockInfo.(*dbInfo.DirBlockInfo)

		receipt.BitcoinTransactionHash = dbi.BTCTxHash.(*primitives.Hash)
		receipt.BitcoinBlockHash = dbi.BTCBlockHash.(*primitives.Hash)
	}

	return receipt, nil
}

func VerifyFullReceipt(dbo interfaces.DBOverlaySimple, receiptStr string) error {
	receipt, err := DecodeReceiptString(receiptStr)
	if err != nil {
		return err
	}

	//fmt.Printf("receipt - %v\n", receipt.CustomMarshalString())

	err = receipt.Validate()
	if err != nil {
		return err
	}

	for i, node := range receipt.MerkleBranch {
		if node.Left == nil || node.Right == nil {
			return fmt.Errorf("Node %v/%v has a nil side", i, len(receipt.MerkleBranch))
		}
		if node.Top == nil {
			return fmt.Errorf("Node %v/%v has no top", i, len(receipt.MerkleBranch))
		}
	}

	return nil
}

func VerifyMinimalReceipt(dbo interfaces.DBOverlaySimple, receiptStr string) error {
	receipt, err := DecodeReceiptString(receiptStr)
	if err != nil {
		return err
	}

	err = receipt.Validate()
	if err != nil {
		return err
	}

	for i, node := range receipt.MerkleBranch {
		if node.Left == nil && node.Right == nil {
			return fmt.Errorf("Node %v/%v has two nil sides", i, len(receipt.MerkleBranch))
		}
		if node.Left != nil && node.Right != nil {
			return fmt.Errorf("Node %v/%v has two non-nil sides", i, len(receipt.MerkleBranch))
		}
		if node.Top != nil {
			return fmt.Errorf("Node %v/%v has unnecessary top", i, len(receipt.MerkleBranch))
		}
	}

	//...

	return nil
}
