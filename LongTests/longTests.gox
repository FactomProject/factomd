package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factom/wallet"
	"github.com/FactomProject/factom/wallet/wsapi"
)

var Seed uint64

const FaucetAddressPriv string = "Fs37iVGnZ7jShPudsXuB98qURxk35eLrmh9cgPuPpTXHAJEBkUTh"
const FaucetAddress string = "FA3Y1tBWnFpyoZUPr9ZH51R1gSC8r5x5kqvkXL3wy4uRvzFnuWLB"

const FactomdWSAPIAddress string = "qatest.factom.org:8088"

//const FactomdWSAPIAddress string = "localhost:8088"

var FAddressStr string
var ECAddressStr string

var ECAddress *factom.ECAddress

var ECKey *primitives.PrivateKey
var FKey *primitives.PrivateKey

func init() {
	fmt.Printf("Run with factomd running!\n")

	// Typically a non-fixed seed should be used, such as .
	// Using a fixed seed will produce the same output on every run.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	Seed = uint64(r.Int63())

	factom.SetFactomdServer(FactomdWSAPIAddress)
}

func InitializeBalance() {
	var once sync.Once
	onceBody := func() {
		w, err := wallet.NewMapDBWallet()
		if err != nil {
			panic(err)
		}

		fa, err := factom.GetFactoidAddress(FaucetAddressPriv)
		err = w.InsertFCTAddress(fa)
		if err != nil {
			panic(err)
		}

		priv, _, _ := testHelper.NewFactoidAddressStrings(Seed)

		FKey, err = primitives.NewPrivateKeyFromHex(priv)
		if err != nil {
			panic(err)
		}
		ECKey, err = primitives.NewPrivateKeyFromHex(priv)
		if err != nil {
			panic(err)
		}

		FAddressStr, err = factoid.PublicKeyStringToFactoidAddressString(FKey.PublicKeyString())
		if err != nil {
			panic(err)
		}

		ECAddressStr, err = factoid.PublicKeyStringToECAddressString(ECKey.PublicKeyString())
		if err != nil {
			panic(err)
		}

		ECAddress, err = factom.MakeECAddress(ECKey.Key[:32])
		if err != nil {
			panic(err)
		}

		go wsapi.Start(w, fmt.Sprintf(":%d", 8089))
		defer func() {
			time.Sleep(10 * time.Millisecond)
			wsapi.Stop()
		}()

		TopupFAddress()
		TopupECAddress()
	}
	once.Do(onceBody)
}

func TopupFAddress() {
	const fBalanceThreshold uint64 = 20000000 //0.2

	fBalance, err := factom.GetFactoidBalance(FaucetAddress)
	if err != nil {
		panic(err)
	}
	if fBalance < int64(fBalanceThreshold) {
		panic(fmt.Sprintf("Balance is too low - %v vs %v", fBalance/100000000.0, fBalanceThreshold/100000000.0))
	}

	tx, err := factom.SendFactoid(FaucetAddress, FAddressStr, fBalanceThreshold, true)
	if err != nil {
		panic(err)
	}

	fmt.Printf("F Topup tx - %v\n", tx)

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx, "")
		if err != nil {
			panic(err)
		}

		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Topup ack - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}

		fmt.Printf("Topup ack - %v\n", str)

		break
	}

	fBalance, _, err = CheckFactomBalance()
	if err != nil {
		panic(err)
	}
	if uint64(fBalance) < fBalanceThreshold {
		panic("Balance was not increased!")
	}

}

func TopupECAddress() {
	//fmt.Printf("TopupECAddress - %v, %v\n", FAddressStr, ECAddressStr)
	const ecBalanceThreshold uint64 = 10

	tx, err := factom.BuyExactEC(FAddressStr, ECAddressStr, uint64(ecBalanceThreshold), true)
	//tx, err := factom.BuyExactEC(FaucetAddress, ECAddressStr, uint64(ecBalanceThreshold), true)
	if err != nil {
		panic(err)
	}

	fmt.Printf("EC Topup tx - %v\n", tx)

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx, "")
		if err != nil {
			panic(err)
		}

		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Topup ack - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}

		fmt.Printf("Topup ack - %v\n", str)

		break
	}

	_, ecBalance, err := CheckFactomBalance()
	if err != nil {
		panic(err)
	}
	if uint64(ecBalance) < ecBalanceThreshold {
		panic("Balance was not increased!")
	}
}

func CheckFactomBalance() (int64, int64, error) {
	ecBalance, err := factom.GetECBalance(ECAddressStr)
	if err != nil {
		return 0, 0, err
	}

	fBalance, err := factom.GetFactoidBalance(FAddressStr)
	if err != nil {
		return 0, 0, err
	}
	return fBalance, ecBalance, nil
}

func main() {
	fmt.Printf("Starting test with seed %v", Seed)
	InitializeBalance()
	CreateChain()
	CreateEntry()
}

func CreateChain() {
	e := CreateNewChain()
	tx1, tx2, err := JustFactomizeChain(e)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created chain %v - %v, %v\n", e.GetChainID(), tx1, tx2)

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx1, "")
		if err != nil {
			panic(err)
		}
		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ack1 - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}
		fmt.Printf("ack1 - %v\n", str)

		break
	}

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx2, "")
		if err != nil {
			panic(err)
		}

		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ack2 - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}

		fmt.Printf("ack2 - %v\n", str)

		break
	}
}

func CreateEntry() {
	e := CreateNewEntry()

	tx1, tx2, err := JustFactomize(e)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created entry %v - %v, %v\n", e.GetChainID(), tx1, tx2)

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx1, "")
		if err != nil {
			panic(err)
		}
		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ack1 - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}
		fmt.Printf("ack1 - %v\n", str)

		break
	}

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx2, "")
		if err != nil {
			panic(err)
		}

		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ack2 - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}

		fmt.Printf("ack2 - %v\n", str)

		break
	}

}

func CreateNewChain() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 0
	answer.ExtIDs = []primitives.ByteSlice{primitives.ByteSlice{Bytes: FKey.Key[:]}}
	answer.Content = primitives.ByteSlice{Bytes: FKey.Public()}
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}

func CreateNewEntry() *entryBlock.Entry {
	chain := CreateNewChain()

	entry := new(entryBlock.Entry)
	entry.ChainID = chain.GetChainID()
	entry.Content = primitives.ByteSlice{Bytes: []byte(FAddressStr)}
	entry.ExtIDs = []primitives.ByteSlice{primitives.ByteSlice{Bytes: []byte(ECAddressStr)}}

	return entry
}

func JustFactomizeChain(entry *entryBlock.Entry) (string, string, error) {
	//Convert entryBlock Entry into factom Entry
	//fmt.Printf("Entry - %v\n", entry)
	j, err := entry.JSONByte()
	if err != nil {
		return "", "", err
	}
	e := new(factom.Entry)
	err = e.UnmarshalJSON(j)
	if err != nil {
		return "", "", err
	}

	chain := factom.NewChain(e)

	//Commit and reveal
	tx1, err := factom.CommitChain(chain, ECAddress)
	if err != nil {
		fmt.Println("Entry commit error : ", err)
		return "", "", err
	}

	time.Sleep(10 * time.Second)
	tx2, err := factom.RevealChain(chain)
	if err != nil {
		fmt.Println("Entry reveal error : ", err)
		return "", "", err
	}

	return tx1, tx2, nil
}

func JustFactomize(entry *entryBlock.Entry) (string, string, error) {
	//Convert entryBlock Entry into factom Entry
	//fmt.Printf("Entry - %v\n", entry)
	j, err := entry.JSONByte()
	if err != nil {
		return "", "", err
	}
	e := new(factom.Entry)
	err = e.UnmarshalJSON(j)
	if err != nil {
		return "", "", err
	}

	//Commit and reveal
	tx1, err := factom.CommitEntry(e, ECAddress)
	if err != nil {
		fmt.Println("Entry commit error : ", err)
		return "", "", err
	}

	time.Sleep(3 * time.Second)
	tx2, err := factom.RevealEntry(e)
	if err != nil {
		fmt.Println("Entry reveal error : ", err)
		return "", "", err
	}

	return tx1, tx2, nil
}
