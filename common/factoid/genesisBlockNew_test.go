package factoid_test

import (
	"encoding/hex"
	"testing"

	"github.com/PaulSnow/factom2d/common/constants"
	. "github.com/PaulSnow/factom2d/common/factoid"
)

func TestGetGenesisFBlockMainNet(t *testing.T) {
	g := GetGenesisFBlock(constants.MAIN_NETWORK_ID)
	if g == nil {
		t.FailNow()
	}
	h, err := g.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	str := hex.EncodeToString(h)
	if str != MainGenesisBlockStr {
		t.Errorf("Wrong binary genesis block!")
	}

	if g.DatabasePrimaryIndex().String() != "a164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f37541736" {
		t.Errorf("Wrong hash")
	}
	if g.DatabaseSecondaryIndex().String() != "2fb170f73c3961d4218ff806dd75e6e348ca1798a5fc7a99d443fbe2ff939d99" {
		t.Errorf("Wrong hash")
	}
}

func TestGetGenesisFBlockExchangeRate(t *testing.T) {
	g := GetGenesisFBlock(constants.MAIN_NETWORK_ID)
	if g == nil {
		t.FailNow()
	}

	if g.GetExchRate() != 666600 {
		t.Errorf("Wrong exchange rate")
	}

	g = GetGenesisFBlock(constants.TEST_NETWORK_ID)
	if g == nil {
		t.FailNow()
	}

	if g.GetExchRate() != 666600 {
		t.Errorf("Wrong exchange rate - %v", g.GetExchRate())
	}

	g = GetGenesisFBlock(0)
	if g == nil {
		t.FailNow()
	}

	if g.GetExchRate() != 1000 {
		t.Errorf("Wrong exchange rate - %v", g.GetExchRate())
	}
}
