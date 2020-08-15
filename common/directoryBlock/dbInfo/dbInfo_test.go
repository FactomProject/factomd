package dbInfo_test

import (
	"bytes"
	"testing"

	. "github.com/PaulSnow/factom2d/common/directoryBlock/dbInfo"
	"github.com/PaulSnow/factom2d/testHelper"
)

func TestMarshalUnmarshal(t *testing.T) {
	var prev *DirBlockInfo = nil
	for i := 0; i < 10; i++ {
		prev = testHelper.CreateTestDirBlockInfo(prev)
		data, err := prev.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		dbi := NewDirBlockInfo()
		err = dbi.UnmarshalBinary(data)
		if err != nil {
			t.Error(err)
		}
		data2, err := dbi.MarshalBinary()

		if bytes.Compare(data, data2) != 0 {
			t.Errorf("Wrong data unmarshalled")
		}
	}
}
