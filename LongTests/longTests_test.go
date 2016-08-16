package longTests_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

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
	const fBalanceThreshold uint64 = 100000000 //1

	tx, err := factom.SendFactoid(FaucetAddress, FAddressStr, fBalanceThreshold)
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

	fBalance, _, err := CheckFactomBalance()
	if err != nil {
		panic(err)
	}
	if uint64(fBalance) < fBalanceThreshold {
		panic("Balance was not increased!")
	}

}

func TopupECAddress() {
	//fmt.Printf("TopupECAddress - %v, %v\n", FAddressStr, ECAddressStr)
	const ecBalanceThreshold uint64 = 100

	tx, err := factom.BuyExactEC(FaucetAddress, FAddressStr, uint64(ecBalanceThreshold))
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

func TestLongTest(t *testing.T) {
	InitializeBalance()
}
