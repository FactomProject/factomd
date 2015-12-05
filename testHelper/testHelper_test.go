package testHelper_test

import (
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

/*
func TestTest(t *testing.T) {
	privKey, pubKey, add := NewFactoidAddressStrings(1)
	t.Errorf("%v, %v, %v", privKey, pubKey, add)
}
*/

func Test(t *testing.T) {
	set := CreateTestBlockSet(nil)
	str, _ := set.ECBlock.JSONString()
	t.Errorf("set ECBlock - %v", str)
	str, _ = set.FBlock.JSONString()
	t.Errorf("set FBlock - %v", str)
	t.Errorf("set Height - %v", set.Height)
}
