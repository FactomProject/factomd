package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestServerFaultMarshalUnmarshal(t *testing.T) {
	sf := new(ServerFault)

	sf.Timestamp = primitives.NewTimestampNow()
	sf.ServerID = testHelper.NewRepeatingHash(1)
	sf.AuditServerID = testHelper.NewRepeatingHash(2)

	sf.VMIndex = 0x33
	sf.DBHeight = 0x44556677
	sf.Height = 0x88990011

	core, err := sf.MarshalCore()
	if err != nil {
		t.Errorf("%v", err)
	}
	for i := 0; i < 10; i++ {
		priv := testHelper.NewPrimitivesPrivateKey(uint64(i))
		sig := priv.Sign(core)
		sf.SignatureList.List = append(sf.SignatureList.List, sig)
	}
	sf.SignatureList.Length = uint32(len(sf.SignatureList.List))

	bin, err := sf.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}

	sf2 := new(ServerFault)
	rest, err := sf2.UnmarshalBinaryData(bin)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) > 0 {
		t.Errorf("Unexpected extra piece of data - %x", rest)
	}
	t.Logf("%v", sf.String())
	t.Logf("%v", sf2.String())

	if sf.Timestamp.GetTimeMilliUInt64() !=
		sf2.Timestamp.GetTimeMilliUInt64() {
		t.Errorf("Invalid Timestamp")
	}
	if sf.ServerID.IsSameAs(sf2.ServerID) == false {
		t.Errorf("Invalid ServerID")
	}
	if sf.AuditServerID.IsSameAs(sf2.AuditServerID) == false {
		t.Errorf("Invalid AuditServerID")
	}
	if sf.VMIndex != sf2.VMIndex {
		t.Errorf("Invalid VMIndex")
	}
	if sf.DBHeight != sf2.DBHeight {
		t.Errorf("Invalid DBHeight")
	}
	if sf.Height != sf2.Height {
		t.Errorf("Invalid Height")
	}

	if sf.SignatureList.Length != sf2.SignatureList.Length {
		t.Errorf("Invalid SignatureList.Length")
	}
	if len(sf.SignatureList.List) != len(sf2.SignatureList.List) {
		t.Errorf("Invalid len of SignatureList.List")
	} else {
		for i := range sf.SignatureList.List {
			if sf.SignatureList.List[i].IsSameAs(sf2.SignatureList.List[i]) == false {
				t.Errorf("Invalid SignatureList.List at %v", i)
			}
		}
	}
}
