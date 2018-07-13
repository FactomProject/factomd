package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/Utilities/tools"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
)

var CheckFloating bool
var UsingAPI bool
var FixIt bool

const level string = "level"
const bolt string = "bolt"

var Debug = false
var Out = "out"

func main() {
	var (
		useApi = flag.Bool("api", false, "Use API instead")
		//addr   = flag.String("addr", "", "Address to check balance of")
	)
	flag.BoolVar(&Debug, "debug", false, "Have debug printing for balances under 0")
	flag.StringVar(&Out, "o", "out", "File to output addresses and balances too")

	flag.Parse()
	UsingAPI = *useApi

	fmt.Println("Usage:")
	fmt.Println("BalanceFinder level/bolt/api DBFileLocation")
	fmt.Println("Program will find balances")

	if len(flag.Args()) < 2 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(flag.Args()) > 2 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	if flag.Args()[0] == "api" {
		UsingAPI = true
	}

	var reader tools.Fetcher

	if UsingAPI {
		reader = tools.NewAPIReader(flag.Args()[1])
	} else {
		levelBolt := flag.Args()[0]

		if levelBolt != level && levelBolt != bolt {
			fmt.Println("\nFirst argument should be `level` or `bolt`")
			os.Exit(1)
		}
		path := flag.Args()[1]
		reader = tools.NewDBReader(levelBolt, path)
	}

	// dblock, err := reader.FetchDBlockHead()

	f, err := os.OpenFile(Out, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
	if err != nil {
		panic(err)
	}

	fct, ec, err := FindBalance(reader)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Addresses and balances written to '%s'\n", Out)

	for k, v := range fct {
		fmt.Fprintf(f, "%s: %d.%d\n", primitives.ConvertFctAddressToUserStr(factoid.NewAddress(k[:])), v/1e8, v%1e8)
	}

	for k, v := range ec {
		fmt.Fprintf(f, "%s: %d\n", primitives.ConvertECAddressToUserStr(factoid.NewAddress(k[:])), v)
	}
}

func FindBalance(reader tools.Fetcher) (map[[32]byte]int64, map[[32]byte]int64, error) {
	top, err := reader.FetchDBlockHead()
	if err != nil {
		return nil, nil, err
	}

	topheight := top.GetDatabaseHeight()

	fctAddressMap := make(map[[32]byte]int64)
	ecAddressMap := make(map[[32]byte]int64)

	for i := uint32(0); i < topheight; i++ {
		//dblock, err := reader.FetchDBlockByHeight(i)
		//if err != nil {
		//	return nil, nil, err
		//}

		if i%1000 == 0 {
			fmt.Printf("Completed %d/%d\n", i, topheight)
		}
		fblock, err := reader.FetchFBlockByHeight(i)
		if err != nil {
			return nil, nil, err
		}

		for _, t := range fblock.GetTransactions() {
			for _, input := range t.GetInputs() {
				fctAddressMap[input.GetAddress().Fixed()] -= int64(input.GetAmount())
			}
			for _, output := range t.GetOutputs() {
				fctAddressMap[output.GetAddress().Fixed()] += int64(output.GetAmount())
			}
			for _, output := range t.GetECOutputs() {
				fctAmt := output.GetAmount()

				ecAddressMap[output.GetAddress().Fixed()] += int64(fctAmt / fblock.GetExchRate())
			}
		}

		ecblock, err := reader.FetchECBlockByHeight(i)
		// ECBlocks 70386-70411 do not exists
		if ecblock == nil && i >= 70386 && i < 70411 {
			continue
			//return nil, nil, fmt.Errorf("ECBlock %d is nil", i)
		}
		for _, entry := range ecblock.GetBody().GetEntries() {
			switch entry.ECID() {
			case constants.ECIDChainCommit:
				ent := entry.(*entryCreditBlock.CommitChain)
				ecAddressMap[ent.ECPubKey.Fixed()] -= int64(ent.Credits)
				DebugIfNeg(ent.ECPubKey.Fixed(), ecAddressMap[ent.ECPubKey.Fixed()], false, i)
			case constants.ECIDEntryCommit:
				ent := entry.(*entryCreditBlock.CommitEntry)
				ecAddressMap[ent.ECPubKey.Fixed()] -= int64(ent.Credits)
				DebugIfNeg(ent.ECPubKey.Fixed(), ecAddressMap[ent.ECPubKey.Fixed()], false, i)
			}
		}
	}
	return fctAddressMap, ecAddressMap, nil
}

func DebugIfNeg(addr [32]byte, amt int64, fct bool, height uint32) {
	if amt < 0 && Debug && height > 97886 {
		str := primitives.ConvertFctAddressToUserStr(factoid.NewAddress(addr[:]))
		if !fct {
			str = primitives.ConvertECAddressToUserStr(factoid.NewAddress(addr[:]))
		}
		fmt.Printf(" -> Balance under 0 : %s with %d funds at blockheight %d\n", str, amt, height)
	}
}
