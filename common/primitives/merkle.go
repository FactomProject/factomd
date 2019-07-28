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

// ComputeMerkleRoot gives a list of hashes, and returns the root of the Merkle Tree
func ComputeMerkleRoot(hashes []interfaces.IHash) interfaces.IHash {
	merkles := BuildMerkleTreeStore(hashes)
	return merkles[len(merkles)-1]
}

// BuildMerkleTreeStore returns the root of the Merkle Tree in merkles[len(merkles)-1]. It interprets
// the input Hash list as the lowest level of the tree, and recursively computes a hash of 2 sequential
// hashs in the input list to generate the next higher level of the tree. If a single hash within a level
// is left over, the function hashs the last hash with itself to produce a final hash of the next level.
// In this manner the next level of the tree is appended to the input hashes until the top (merkle root)
// is appended last.
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

// MerkleNode is a node in a Merkle tree
type MerkleNode struct {
	Left  *Hash `json:"left,omitempty"`  // Left tree node hash
	Right *Hash `json:"right,omitempty"` // Right tree node hash
	Top   *Hash `json:"top,omitempty"`   // Top tree node hash
}

// BuildMerkleBranchForHash creates a Merkle tree from an input list of hashs and picks out a single branch
// of that tree starting with the target hash and traversing up the Merkle tree until the
// Merkle root node at the end of the returned array. By default only the compliment node information
// is stored in each node of the branch, but the 'fullDetail' option stores the compliment, self, and top
// nodes within each Merkle node.
func BuildMerkleBranchForHash(hashes []interfaces.IHash, target interfaces.IHash, fullDetail bool) []*MerkleNode {
	for i, h := range hashes {
		if h.IsSameAs(target) {
			return BuildMerkleBranch(hashes, i, fullDetail)
		}
	}
	return nil
}

// BuildMerkleBranch creates a Merkle tree from an input list of hashs and picks out a single branch
// of that tree starting with the 'entryIndex' node and traversing up the Merkle tree until the
// Merkle root node at the end of the returned array. By default only the compliment node information
// is stored in each node of the branch, but the 'fullDetail' option stores the compliment, self, and top
// nodes within each Merkle node.
func BuildMerkleBranch(hashes []interfaces.IHash, entryIndex int, fullDetail bool) []*MerkleNode {
	if len(hashes) < entryIndex || len(hashes) == 0 {
		return nil
	}
	merkleTree := BuildMerkleTreeStore(hashes)
	//fmt.Printf("Merkle tree - %v\n", merkleTree)
	levelWidth := len(hashes) // base tree width
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
		if index%2 == 0 { // Node of interest is a left node
			complimentIndex = index + 1        // corresponding right node index
			if complimentIndex == levelWidth { // if right node index takes us past the edge of the base width then
				complimentIndex = index // the compliment index must be itself (single node)
			}
			topIndex = index/2 + levelWidth // Top node is the base level width plus half the index position (because its a binary tree)
			// Store the neighboring node information in the MerkleNode
			mn.Right = merkleTree[offset+complimentIndex].(*Hash)
			if fullDetail == true {
				mn.Left = merkleTree[offset+index].(*Hash)
				mn.Top = merkleTree[offset+topIndex].(*Hash)
			}
		} else { // Node of interest is a right node
			complimentIndex = index - 1 // corresponding left node index
			topIndex = complimentIndex/2 + levelWidth
			// Store the neighboring node information in the MerkleNode
			mn.Left = merkleTree[offset+complimentIndex].(*Hash)
			if fullDetail == true {
				mn.Right = merkleTree[offset+index].(*Hash)
				mn.Top = merkleTree[offset+topIndex].(*Hash)
			}
		}
		// Save node of interest
		answer = append(answer, mn)

		// Move up to the next level above, the topIndex location
		offset += levelWidth
		index = topIndex - levelWidth
		levelWidth = (levelWidth + 1) / 2
	}
	return answer
}
