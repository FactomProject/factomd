// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"fmt"
	"math"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

var _ = fmt.Println

// NextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two.  This is a helper function used during the
// calculation of a merkle tree.
func NextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

// HashMerkleBranches takes two hashes, treated as the left and right tree
// nodes, and returns the hash of their concatenation.  This is a helper
// function used to aid in the generation of a merkle tree.
func HashMerkleBranches(left interfaces.IHash, right interfaces.IHash) interfaces.IHash {
	// Concatenate the left and right nodes.
	var barray []byte = make([]byte, constants.ADDRESS_LENGTH*2)
	copy(barray[:constants.ADDRESS_LENGTH], left.Bytes())
	copy(barray[constants.ADDRESS_LENGTH:], right.Bytes())

	newSha := Sha(barray)
	return newSha
}

// Give a list of hashes, return the root of the Merkle Tree
func ComputeMerkleRoot(hashes []interfaces.IHash) interfaces.IHash {
	merkles := BuildMerkleTreeStore(hashes)
	return merkles[len(merkles)-1]
}

// The root of the Merkle Tree is returned in merkles[len(merkles)-1]
func BuildMerkleTreeStore(hashes []interfaces.IHash) (merkles []interfaces.IHash) {
	if len(hashes) == 0 {
		return append(make([]interfaces.IHash, 0, 1), new(Hash))
	}
	if len(hashes) < 2 {
		return hashes
	}
	nextLevel := []interfaces.IHash{}
	for i := 0; i < len(hashes); i += 2 {
		var node interfaces.IHash
		if i+1 == len(hashes) {
			node = HashMerkleBranches(hashes[i], hashes[i])
		} else {
			node = HashMerkleBranches(hashes[i], hashes[i+1])
		}
		nextLevel = append(nextLevel, node)
	}
	nextIteration := BuildMerkleTreeStore(nextLevel)
	return append(hashes, nextIteration...)
}

type MerkleNode struct {
	Left  *Hash `json:"left,omitempty"`
	Right *Hash `json:"right,omitempty"`
	Top   *Hash `json:"top,omitempty"`
}

func BuildMerkleBranchForHash(hashes []interfaces.IHash, target interfaces.IHash, fullDetail bool) []*MerkleNode {
	for i, h := range hashes {
		if h.IsSameAs(target) {
			return BuildMerkleBranch(hashes, i, fullDetail)
		}
	}
	return nil
}

func BuildMerkleBranch(hashes []interfaces.IHash, entryIndex int, fullDetail bool) []*MerkleNode {
	if len(hashes) < entryIndex || len(hashes) == 0 {
		return nil
	}
	merkleTree := BuildMerkleTreeStore(hashes)
	//fmt.Printf("Merkle tree - %v\n", merkleTree)
	levelWidth := len(hashes)
	complimentIndex := 0
	topIndex := 0
	index := entryIndex
	answer := []*MerkleNode{}
	offset := 0
	for {
		/*fmt.Printf("Index %v out of %v\n", offset+index, len(merkleTree))
		fmt.Printf("levelWidth - %v\n", levelWidth)
		fmt.Printf("offset - %v\n", offset)*/
		if levelWidth == 1 {
			break
		}
		mn := new(MerkleNode)
		if index%2 == 0 {
			complimentIndex = index + 1
			if complimentIndex == levelWidth {
				complimentIndex = index
			}
			topIndex = index/2 + levelWidth
			mn.Right = merkleTree[offset+complimentIndex].(*Hash)
			if fullDetail == true {
				mn.Left = merkleTree[offset+index].(*Hash)
				mn.Top = merkleTree[offset+topIndex].(*Hash)
			}
		} else {
			complimentIndex = index - 1
			topIndex = complimentIndex/2 + levelWidth
			mn.Left = merkleTree[offset+complimentIndex].(*Hash)
			if fullDetail == true {
				mn.Right = merkleTree[offset+index].(*Hash)
				mn.Top = merkleTree[offset+topIndex].(*Hash)
			}
		}
		answer = append(answer, mn)
		offset += levelWidth
		index = topIndex - levelWidth
		levelWidth = (levelWidth + 1) / 2
	}
	return answer
}
