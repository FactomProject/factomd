package main

import (
	"errors"
	"flag"

	"fmt"

	"os"

	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

func main() {
	var (
		filename = flag.String("f", "FastBoot_MAIN_v8.db", "FastbootFile location")
	)

	flag.Parse()

	s := testHelper.CreateEmptyTestState()

	statelist := s.DBStates
	fmt.Println(statelist.State.FactomNodeName, "Loading from", filename)
	b, err := state.LoadFromFile(s, *filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "LoadDBStateList error:", err)
		panic(err)
	}
	if b == nil {
		fmt.Fprintln(os.Stderr, "LoadDBStateList LoadFromFile returned nil")
		panic(errors.New("failed to load from file"))
	}
	h := primitives.NewZeroHash()
	b, err = h.UnmarshalBinaryData(b)
	if err != nil {
		panic(err)
	}
	h2 := primitives.Sha(b)
	if h.IsSameAs(h2) == false {
		fmt.Fprintf(os.Stderr, "LoadDBStateList - Integrity hashes do not match!")
		panic(errors.New("fastboot file does not match its hash"))
		//return fmt.Errorf("Integrity hashes do not match")
	}

	statelist.UnmarshalBinary(b)
	var i int
	for i = len(statelist.DBStates) - 1; i >= 0; i-- {
		if statelist.DBStates[i].SaveStruct != nil {
			break
		}
	}
	statelist.DBStates[i].SaveStruct.RestoreFactomdState(statelist.State)

	//fmt.Println(s.IdentityControl)

	state.PrintState(s)

	h1 := state.GetMapHash(s.FactoidBalancesP)
	h2 = state.GetMapHash(s.ECBalancesP)

	var d []byte
	d = append(d, h1.Bytes()...)
	d = append(d, h2.Bytes()...)
	r := primitives.Sha(b)

	fmt.Printf("Balance Hash: DBHeight %d, FCTCount %d, ECCount %d, Hash %x\n", s.GetLLeaderHeight(), len(s.FactoidBalancesP), len(s.ECBalancesP), r.Bytes()[:])

	s.FactoidState.(*state.FactoidState).DBHeight = s.GetLLeaderHeight()
	bh := s.FactoidState.GetBalanceHash(false)
	fmt.Printf("-- State --\n"+
		"Height: %d\n"+
		"FCT Address Count: %d\n"+
		"EC Address Count: %d\n"+
		"Balance Hash: %s\n",
		s.LLeaderHeight, len(s.FactoidBalancesP), len(s.ECBalancesP), bh.String())

	// Identity Related Info
	fmt.Println(s.IdentityControl)
}
