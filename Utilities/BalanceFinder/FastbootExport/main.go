package main

import (
	"flag"
	"os"

	"io/ioutil"

	"fmt"

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
	//dbs := new(state.DBStateList)
	file, err := os.OpenFile(*filename, os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	h := primitives.NewZeroHash()
	data, err = h.UnmarshalBinaryData(data)
	if err != nil {
		panic(err)
	}
	h2 := primitives.Sha(data)
	if h.IsSameAs(h2) == false {
		fmt.Printf("LoadDBStateList - Integrity hashes do not match!")
		panic(err)
		//return fmt.Errorf("Integrity hashes do not match")
	}

	nd, err := s.DBStates.UnmarshalBinaryData(data)
	if err != nil {
		panic(err)
	}

	if len(nd) != 0 {
		panic("Left over bytes after savestate unmarshal")
	}

	state.PrintState(s)
	b := s.FactoidState.GetBalanceHash(false)
	fmt.Printf("-- State --\n"+
		"Height: %d\n"+
		"Balance Hash: %s\n",
		s.LLeaderHeight, b.String())
}
