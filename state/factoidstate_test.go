// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

var fs interfaces.IFactoidState

func RandBal() int64 {
	switch rand.Int() & 7 {
	case 0:
		return rand.Int63()
	case 1:
		return rand.Int63() >> 8
	case 2:
		return rand.Int63() >> 16
	case 3:
		return rand.Int63() >> 24
	case 4:
		return rand.Int63() >> 32
	case 5:
		return rand.Int63() >> 40
	case 6:
		return rand.Int63() >> 48
	case 7:
		return rand.Int63() >> 56
	}
	return 0
}

func RandBit() (bit int64) {
	bit = 1
	bit = bit << uint32(rand.Int()%64)
	return
}

func TestBalanceHash(t *testing.T) {
	s := new(State)
	fs := new(FactoidState)
	s.FactoidState = fs
	fs.State = s
	s.FactoidBalancesP = map[[32]byte]int64{}
	s.ECBalancesP = map[[32]byte]int64{}

	var ec, fct []interfaces.IHash
	h := primitives.Sha([]byte("testing"))

	for i := 1; i < 1000; i++ {
		h = primitives.Sha(h.Bytes())
		ec = append(ec, h)
		s.PutE(false, h.Fixed(), RandBal())
		h = primitives.Sha(h.Bytes())
		fct = append(fct, h)
		s.PutF(false, h.Fixed(), RandBal())
	}

	Expected := fs.GetBalanceHash(false).String()
	hbal := fs.GetBalanceHash(false)

	if hbal.String() != Expected {
		t.Errorf("Expected %s but found %s", Expected, hbal.String())
	}

	x := func(addrArray []interfaces.IHash, balanceArray *map[[32]byte]int64) {

		// Add a random address
		for i := 1; i < 10; i++ {
			h = primitives.Sha(h.Bytes())
			adr := h
			bal := RandBal()
			(*balanceArray)[adr.Fixed()] = bal

			hbal := fs.GetBalanceHash(false)

			if hbal.String() == Expected {
				t.Errorf("Should not have gotten %s", Expected)
			}

			delete((*balanceArray), adr.Fixed())

			hbal = fs.GetBalanceHash(false)

			if hbal.String() != Expected {
				t.Errorf("Expected %s but found %s", Expected, hbal.String())
			}
		}

		// Delete a random address
		for i := 1; i < 10; i++ {
			indx := rand.Int() % len(addrArray)
			adr := addrArray[indx].Fixed()
			bal := (*balanceArray)[adr]
			delete((*balanceArray), adr)

			hbal := fs.GetBalanceHash(false)

			if hbal.String() == Expected {
				t.Errorf("Should not have gotten %s", Expected)
			}
			(*balanceArray)[adr] = bal

			hbal = fs.GetBalanceHash(false)

			if hbal.String() != Expected {
				t.Errorf("Expected %s but found %s", Expected, hbal.String())
			}
		}

		// Modify by one bit a random balance
		for i := 1; i < 10; i++ {
			indx := rand.Int() % len(addrArray)
			adr := addrArray[indx].Fixed()

			bal := (*balanceArray)[adr]
			(*balanceArray)[adr] = bal ^ RandBit()

			hbal := fs.GetBalanceHash(false)
			if hbal.String() == Expected {
				t.Errorf("Should not have gotten %s", Expected)
			}

			(*balanceArray)[adr] = bal

			hbal = fs.GetBalanceHash(false)

			if hbal.String() != Expected {
				t.Errorf("Expected %s but found %s", Expected, hbal.String())
			}

		}

	}

	x(fct, &s.FactoidBalancesP)
	x(ec, &s.ECBalancesP)

}

func TestGetMapHash(t *testing.T) {
	bmap := map[[32]byte]int64{}

	//using some arbitrary IDs
	h, _ := primitives.NewShaHash(constants.EC_CHAINID)
	bmap[h.Fixed()] = 0
	h, _ = primitives.NewShaHash(constants.D_CHAINID)
	bmap[h.Fixed()] = 1
	h, _ = primitives.NewShaHash(constants.ADMIN_CHAINID)
	bmap[h.Fixed()] = math.MaxInt64
	h, _ = primitives.NewShaHash(constants.FACTOID_CHAINID)
	bmap[h.Fixed()] = math.MinInt64
	h, _ = primitives.NewShaHash(constants.ZERO_HASH)
	bmap[h.Fixed()] = 123456789

	h2 := GetMapHash(bmap)
	if h2 == nil {
		t.Errorf("Hot nil hash")
	}
	// expected := "fd9b4c42a47115af0bf1878c7de793e28b021415f82ed7151ab0cbb7db941b31" The expected valuie changes when we quite including the height
	expected := "b26a603681665b05acaa37627d037d8e6cb23e19161affedb3fd09f283494024"

	s := h2.String()
	if s != expected {
		t.Errorf("Invalid hash - got %v, expected %v", s, expected)
	}

	for i := 0; i < 1000; i++ {
		bmap = map[[32]byte]int64{}
		l := random.RandIntBetween(0, 100)
		for j := 0; j < l; j++ {
			bmap[primitives.RandomHash().Fixed()] = random.RandInt64()
		}
		h2 = GetMapHash(bmap)
		for j := 0; j < 10; j++ {
			if h2.IsSameAs(GetMapHash(bmap)) == false {
				t.Errorf("GetMapHash returns inconsistent hashes")
			}
		}
	}
}

func TestUpdateECTransactionWithNegativeBalance(t *testing.T) {
	s := testHelper.CreateAndPopulateTestStateAndStartValidator()
	fs := s.FactoidState

	add1, err := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Error(err)
		return
	}

	s.PutE(true, add1.Fixed(), 10)

	add1bs := primitives.StringToByteSlice32("0000000000000000000000000000000000000000000000000000000000000001")
	cc := new(entryCreditBlock.CommitChain)
	cc.ECPubKey = add1bs
	cc.Credits = 10

	err = fs.UpdateECTransaction(true, cc)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = fs.UpdateECTransaction(true, cc)
	if err == nil {
		t.Errorf("No error returned when it should be")
	}
	s.NetworkNumber = constants.NETWORK_MAIN
	err = fs.UpdateECTransaction(true, cc)
	if err != nil {
		t.Errorf("%v", err)
	}

	fs.(*FactoidState).DBHeight = 97887
	err = fs.UpdateECTransaction(true, cc)
	if err == nil {
		t.Errorf("No error returned when it should be")
	}

	s = testHelper.CreateAndPopulateTestStateAndStartValidator()
	fs = s.FactoidState
	s.PutE(true, add1.Fixed(), 10)

	ce := new(entryCreditBlock.CommitEntry)
	ce.ECPubKey = add1bs
	ce.Credits = 10

	err = fs.UpdateECTransaction(true, ce)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = fs.UpdateECTransaction(true, ce)
	if err == nil {
		t.Errorf("No error returned when it should be")
	}
	s.NetworkNumber = constants.NETWORK_MAIN
	err = fs.UpdateECTransaction(true, ce)
	if err != nil {
		t.Errorf("%v", err)
	}
	fs.(*FactoidState).DBHeight = 97887
	err = fs.UpdateECTransaction(true, ce)
	if err == nil {
		t.Errorf("No error returned when it should be")
	}

	s.PutE(true, add1.Fixed(), 0)
}

func TestUpdateTransactionWithNegativeBalance(t *testing.T) {
	s := testHelper.CreateAndPopulateTestStateAndStartValidator()
	fs := s.FactoidState

	add1, err := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Error(err)
		return
	}

	s.PutF(true, add1.Fixed(), 10)

	//add1bs := primitives.StringToByteSlice32("0000000000000000000000000000000000000000000000000000000000000001")

	ft := new(factoid.Transaction)
	ta := new(factoid.TransAddress)
	ta.Address = add1
	ta.Amount = 10
	ta.UserAddress = "abc"
	ft.Inputs = []interfaces.ITransAddress{ta}

	err = fs.UpdateTransaction(true, ft)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = fs.UpdateTransaction(true, ft)
	if err == nil {
		t.Errorf("No error returned when it should be")
	}
}

/*
func TestUpdateECTransaction(t *testing.T) {
	fs.SetFactoshisPerEC(1)
	add1, err := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Error(err)
		return
	}
	add1bs := primitives.StringToByteSlice32("0000000000000000000000000000000000000000000000000000000000000001")

	if fs.GetECBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
		return
	}

	var tx interfaces.IECBlockEntry
	tx = new(entryCreditBlock.ServerIndexNumber)

	err = fs.UpdateECTransaction(tx)
	if err != nil {
		t.Error(err)
		return
	}
	if fs.GetECBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
	}

	tx = new(entryCreditBlock.MinuteNumber)

	err = fs.UpdateECTransaction(tx)
	if err != nil {
		t.Error(err)
		return
	}
	if fs.GetECBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
		return
	}

	//Proper processing
	cc := new(entryCreditBlock.CommitChain)
	cc.ECPubKey = add1bs
	cc.Credits = 100
	tx = cc

	err = fs.UpdateECTransaction(tx)
	if err != nil {
		t.Error(err)
		return
	}
	if fs.GetECBalance(add1.Fixed()) != -100 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
		return
	}

	ib := new(entryCreditBlock.IncreaseBalance)
	ib.ECPubKey = add1bs
	ib.NumEC = 100
	tx = ib

	err = fs.UpdateECTransaction(tx)
	if err != nil {
		t.Error(err)
		return
	}
	if fs.GetECBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
		return
	}

	ce := new(entryCreditBlock.CommitEntry)
	ce.ECPubKey = add1bs
	ce.Credits = 100
	tx = ce

	err = fs.UpdateECTransaction(tx)
	if err != nil {
		t.Error(err)
		return
	}
	if fs.GetECBalance(add1.Fixed()) != -100 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
		return
	}

}
*/

/*
func TestBalances(t *testing.T) {
	s := testHelper.CreateEmptyTestState()
	fs = s.GetFactoidState()
	fs.SetFactoshisPerEC(1)
	add1, err := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Error(err)
	}
	add2, err := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000002")
	if err != nil {
		t.Error(err)
	}
	add3, err := primitives.HexToHash("0000000000000000000000000000000000000000000000000000000000000003")
	if err != nil {
		t.Error(err)
	}

	tx := new(factoid.Transaction)
	tx.AddOutput(add1, 1000000)

	err = fs.UpdateTransaction(tx)
	if err != nil {
		t.Error(err)
	}

	if fs.GetFactoidBalance(add1.Fixed()) != 1000000 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add1.Fixed()))
	}
	if fs.GetECBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
	}
	if fs.GetFactoidBalance(add2.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add2.Fixed()))
	}
	if fs.GetECBalance(add2.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add2.Fixed()))
	}
	if fs.GetFactoidBalance(add3.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add3.Fixed()))
	}
	if fs.GetECBalance(add3.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add3.Fixed()))
	}

	tx = new(factoid.Transaction)
	tx.AddInput(add1, 1000)
	tx.AddOutput(add2, 1000)

	err = fs.UpdateTransaction(tx)
	if err != nil {
		t.Error(err)
	}

	if fs.GetFactoidBalance(add1.Fixed()) != 999000 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add1.Fixed()))
	}
	if fs.GetECBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
	}
	if fs.GetFactoidBalance(add2.Fixed()) != 1000 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add2.Fixed()))
	}
	if fs.GetECBalance(add2.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add2.Fixed()))
	}
	if fs.GetFactoidBalance(add3.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add3.Fixed()))
	}
	if fs.GetECBalance(add3.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add3.Fixed()))
	}

	tx = new(factoid.Transaction)
	tx.AddInput(add1, 1000)
	tx.AddECOutput(add3, 1000)

	err = fs.UpdateTransaction(tx)
	if err != nil {
		t.Error(err)
	}

	if fs.GetFactoidBalance(add1.Fixed()) != 998000 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add1.Fixed()))
	}
	if fs.GetECBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
	}
	if fs.GetFactoidBalance(add2.Fixed()) != 1000 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add2.Fixed()))
	}
	if fs.GetECBalance(add2.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add2.Fixed()))
	}
	if fs.GetFactoidBalance(add3.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add3.Fixed()))
	}
	if fs.GetECBalance(add3.Fixed()) != 1000 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add3.Fixed()))
	}

	fs.ResetBalances()

	if fs.GetFactoidBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add1.Fixed()))
	}
	if fs.GetECBalance(add1.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add1.Fixed()))
	}
	if fs.GetFactoidBalance(add2.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add2.Fixed()))
	}
	if fs.GetECBalance(add2.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add2.Fixed()))
	}
	if fs.GetFactoidBalance(add3.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetFactoidBalance(add3.Fixed()))
	}
	if fs.GetECBalance(add3.Fixed()) != 0 {
		t.Errorf("Invalid address balance - %v", fs.GetECBalance(add3.Fixed()))
	}
}



import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/boltdb"
	"math/rand"
	"testing"
)

var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = Prtln

func GetDatabase() interfaces. {
	var bucketList [][]byte

	bucketList = make([][]byte, 5, 5)

	bucketList[0] = []byte("factoidAddress_balances")
	bucketList[0] = []byte("factoidOrphans_balances")
	bucketList[0] = []byte("factomAddress_balances")

	db := new(BoltDB)

	db.Init(bucketList, "/tmp/fs_test.db")

	return db
}

func Test_updating_balances_FactoidState(test *testing.T) {
	fs := new(FactoidState)
	fs.database = GetDatabase()

}
*/
