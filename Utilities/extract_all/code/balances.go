package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/FactomProject/factomd/state"
)

type account struct {
	address [32]byte
	balance int64
}

func (b *account) MarshalBinary() []byte {
	var data [40]byte
	copy(data[:32], b.address[:])
	binary.BigEndian.PutUint64(data[32:], uint64(b.balance))
	return data[:]
}

func (b *account) MarshalText() []byte {
	s := fmt.Sprintf("%032x %d.%08d\n", b.address, b.balance/100000000, b.balance%100000000)
	return []byte(s)
}

func (b *account) MarshalTextEC() []byte {
	s := fmt.Sprintf("%032x %d\n", b.address, b.balance)
	return []byte(s)
}

func ProcessBalances() {
	state := FactomdState.(*state.State)       // Get the Factom State
	state.FactoidBalancesPMutex.Lock()         // Lock the balances
	defer state.FactoidBalancesPMutex.Unlock() // unlock the balances when done

	var factoids []account     // Collect Factoids
	var entryCredits []account // Collect Entry Credits

	for k, v := range state.FactoidBalancesP { // Pull all the Factoid balances
		if v != 0 {
			b := account{address: k, balance: v}
			factoids = append(factoids, b)
		}
	}

	for k, v := range state.ECBalancesP { // Pull all the Entry Credit balances
		if v != 0 {
			b := account{address: k, balance: v}
			entryCredits = append(entryCredits, b)
		}
	}

	sort.Slice(factoids, func(i, j int) bool { // Sort the Factoids by address
		return bytes.Compare(factoids[i].address[:], factoids[j].address[:]) < 0
	})

	sort.Slice(entryCredits, func(i, j int) bool { // Sort the Entry Credits by address
		return bytes.Compare(entryCredits[i].address[:], entryCredits[j].address[:]) < 0
	})

	// Write out binary balances
	filename := "balances.dat"
	if f, err := os.Create(path.Join(FullDir, filename)); err != nil {
		panic(fmt.Sprintf("Could not open %s: %v", path.Join(FullDir, filename), err))
	} else {
		OutputFile = f
		defer OutputFile.Close()
	}

	h := Header{Tag: TagFCT, Size: 40}
	hb := h.MarshalBinary()
	for _, v := range factoids {
		OutputFile.Write(hb)
		OutputFile.Write(v.MarshalBinary())

	}

	h = Header{Tag: TagEC, Size: 40}
	hb = h.MarshalBinary()
	for _, v := range entryCredits {
		OutputFile.Write(hb)
		OutputFile.Write(v.MarshalBinary())

	}

	// Write out text balances
	filename = "balances.txt"
	if f, err := os.Create(path.Join(FullDir, filename)); err != nil {
		panic(fmt.Sprintf("Could not open %s: %v", path.Join(FullDir, filename), err))
	} else {
		OutputFile = f
	}

	for _, v := range factoids {
		OutputFile.Write([]byte("FCT: "))
		OutputFile.Write(v.MarshalText())
	}

	for _, v := range entryCredits {
		OutputFile.Write([]byte("EC:  "))
		OutputFile.Write(v.MarshalTextEC())
	}

	FCTAccountCnt = uint64(len(factoids))
	ECAccountCnt = uint64(len(entryCredits))

}
