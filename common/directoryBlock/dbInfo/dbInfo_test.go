package dbInfo_test

import (
	. "github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	dbi := NewDirBlockInfo()
	data, err := dbi.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	err = dbi.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
	}

}
