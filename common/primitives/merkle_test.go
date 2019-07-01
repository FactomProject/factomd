// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

func TestNextPowerOfTwo(t *testing.T) {
	if NextPowerOfTwo(0) != 0 {
		t.Error("TestNextPowerOfTwo failed for 0")
	}
	if NextPowerOfTwo(1) != 1 {
		t.Error("TestNextPowerOfTwo failed for 1")
	}
	if NextPowerOfTwo(2) != 2 {
		t.Error("TestNextPowerOfTwo failed for 2")
	}
	if NextPowerOfTwo(3) != 4 {
		t.Error("TestNextPowerOfTwo failed for 3")
	}
	if NextPowerOfTwo(4) != 4 {
		t.Error("TestNextPowerOfTwo failed for 4")
	}
	if NextPowerOfTwo(5) != 8 {
		t.Error("TestNextPowerOfTwo failed for 5")
	}
	if NextPowerOfTwo(6) != 8 {
		t.Error("TestNextPowerOfTwo failed for 6")
	}
	if NextPowerOfTwo(7) != 8 {
		t.Error("TestNextPowerOfTwo failed for 7")
	}
	if NextPowerOfTwo(8) != 8 {
		t.Error("TestNextPowerOfTwo failed for 8")
	}
}

func TestHashMerkleBranches(t *testing.T) {
	h1, err := NewShaHashFromStr("82501c1178fa0b222c1f3d474ec726b832013f0a532b44bb620cce8624a5feb1")
	if err != nil {
		t.Error(err)
	}
	h2, err := NewShaHashFromStr("169e1e83e930853391bc6f35f605c6754cfead57cf8387639d3b4096c54f18f4")
	if err != nil {
		t.Error(err)
	}
	root, err := NewShaHashFromStr("a24ee7fb7333f85c16560ed8850a1773d6977ce7a4936367eaf72f8fff33797e")
	if err != nil {
		t.Error(err)
	}

	answer := HashMerkleBranches(h1, h2)

	if answer.IsSameAs(root) == false {
		t.Errorf("TestHashMerkleBranches failed - Received %v, expected %v", answer, root)
	}
}

func TestBuildMerkleTreeStore(t *testing.T) {
	max := 9
	list := buildMerkleLeafs(max)
	merkles := BuildMerkleTreeStore(list)
	expected := buildExpectedMerkleTree(list)
	t.Logf("merkles - %v", merkles)
	t.Logf("expected - %v", expected)

	if len(merkles) != len(expected) {
		t.Logf("lends are not identical - %v vs %v", len(merkles), len(expected))
	}

	for i := 0; i < len(merkles); i++ {
		if merkles[i] == nil {
			t.Errorf("Merkle %v/%v is nil!", i, len(merkles))
		} else {
			if merkles[i].IsSameAs(expected[i]) == false {
				t.Errorf("Hash %v/%v is not equal - %v vs %v", i, len(merkles), merkles[i], expected[i])
			}
		}
	}
}

func TestBuildMerkleBranch(t *testing.T) {
	max := 9
	list := buildMerkleLeafs(max)
	tree := buildExpectedMerkleTree(list)
	branch := BuildMerkleBranch(list, max-1, true)

	leftIndexes := []int{8, 13, 16, 17}
	rightIndexes := []int{8, 13, 16, 18}
	topIndexes := []int{13, 16, 18, 19}

	for i := 0; i < 4; i++ {
		if branch[i].Left.IsSameAs(tree[leftIndexes[i]]) == false {
			t.Errorf("Left node is wrong on index %v", i)
		}
		if branch[i].Right.IsSameAs(tree[rightIndexes[i]]) == false {
			t.Errorf("Right node is wrong on index %v", i)
		}
		if branch[i].Top.IsSameAs(tree[topIndexes[i]]) == false {
			t.Errorf("Top node is wrong on index %v", i)
		}
	}

	branch = BuildMerkleBranch(list, 0, true)

	leftIndexes = []int{0, 9, 14, 17}
	rightIndexes = []int{1, 10, 15, 18}
	topIndexes = []int{9, 14, 17, 19}

	for i := 0; i < 4; i++ {
		if branch[i].Left.IsSameAs(tree[leftIndexes[i]]) == false {
			t.Errorf("Left node is wrong on index %v", i)
		}
		if branch[i].Right.IsSameAs(tree[rightIndexes[i]]) == false {
			t.Errorf("Right node is wrong on index %v", i)
		}
		if branch[i].Top.IsSameAs(tree[topIndexes[i]]) == false {
			t.Errorf("Top node is wrong on index %v", i)
		}
	}

	branch = BuildMerkleBranch(list, 6, true)

	leftIndexes = []int{6, 11, 14, 17}
	rightIndexes = []int{7, 12, 15, 18}
	topIndexes = []int{12, 15, 17, 19}

	for i := 0; i < 4; i++ {
		if branch[i].Left.IsSameAs(tree[leftIndexes[i]]) == false {
			t.Errorf("Left node is wrong on index %v", i)
		}
		if branch[i].Right.IsSameAs(tree[rightIndexes[i]]) == false {
			t.Errorf("Right node is wrong on index %v", i)
		}
		if branch[i].Top.IsSameAs(tree[topIndexes[i]]) == false {
			t.Errorf("Top node is wrong on index %v", i)
		}
	}
}

func generateHash(n int) interfaces.IHash {
	answer := ""
	for i := 0; i < 64; i++ {
		answer = answer + fmt.Sprintf("%v", n)
	}
	hash, err := NewShaHashFromStr(answer)
	if err != nil {
		panic(err)
	}
	return hash
}

func buildMerkleLeafs(n int) []interfaces.IHash {
	list := make([]interfaces.IHash, n)
	for i := 0; i < n; i++ {
		list[i] = generateHash(i)
	}
	return list
}

func buildExpectedMerkleTree(hashes []interfaces.IHash) []interfaces.IHash {
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
	nextIteration := buildExpectedMerkleTree(nextLevel)
	return append(hashes, nextIteration...)
}
