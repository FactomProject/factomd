package adminBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestServerFaultGetHash(t *testing.T) {
	a := new(ServerFault)
	h := a.Hash()
	expected := "5039b1b0a2a8420f89bfc5527c2c8b596a3f7c49d05eb70a53d821668523c9b8"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestServerFaultTypeIDCheck(t *testing.T) {
	a := new(ServerFault)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(ServerFault)
	err = a2.UnmarshalBinary(b)
	if err != nil {
		t.Errorf("%v", err)
	}

	b[0] = (b[0] + 1) % 255
	err = a2.UnmarshalBinary(b)
	if err == nil {
		t.Errorf("No error caught")
	}
}

func TestUnmarshalNilServerFault(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ServerFault)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

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

func TestFaultMisc(t *testing.T) {
	a := new(ServerFault)
	if a.String() != "    E:                    EntryServerFault -- DBheight 0s ServerID   000000 AuditServer     0000, #sigs 0, VMIndex 0" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":10,\"timestamp\":0,\"serverid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"auditserverid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"vmindex\":0,\"dbheight\":0,\"height\":0,\"signaturelist\":{\"Length\":0,\"List\":null}}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":10,\"timestamp\":0,\"serverid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"auditserverid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"vmindex\":0,\"dbheight\":0,\"height\":0,\"signaturelist\":{\"Length\":0,\"List\":null}}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}

// This test always fails
// func TestServerFaultUpdateState(t *testing.T) {
// 	sigs := 10
// 	sf := new(ServerFault)

// 	sf.Timestamp = primitives.NewTimestampNow()
// 	sf.ServerID = testHelper.NewRepeatingHash(1)
// 	sf.AuditServerID = testHelper.NewRepeatingHash(2)

// 	sf.VMIndex = 0x33
// 	sf.DBHeight = 0x44556677
// 	sf.Height = 0x88990011

// 	core, err := sf.MarshalCore()
// 	if err != nil {
// 		t.Errorf("%v", err)
// 	}
// 	for i := 0; i < sigs; i++ {
// 		priv := testHelper.NewPrimitivesPrivateKey(uint64(i))
// 		sig := priv.Sign(core)
// 		sf.SignatureList.List = append(sf.SignatureList.List, sig)
// 	}
// 	sf.SignatureList.Length = uint32(len(sf.SignatureList.List))

// 	s := testHelper.CreateAndPopulateTestState()
// 	idindex := s.CreateBlankFactomIdentity(primitives.NewZeroHash())
// 	s.Identities[idindex].ManagementChainID = primitives.NewZeroHash()
// 	for i := 0; i < sigs; i++ {
// 		//Federated Server
// 		index := s.AddAuthorityFromChainID(testHelper.NewRepeatingHash(byte(i)))
// 		s.Authorities[index].SigningKey = *testHelper.NewPrimitivesPrivateKey(uint64(i)).Pub
// 		s.Authorities[index].Status = 1

// 		s.AddFedServer(s.GetLeaderHeight(), testHelper.NewRepeatingHash(byte(i)))

// 		//Audit Server
// 		index = s.AddAuthorityFromChainID(testHelper.NewRepeatingHash(byte(i + sigs)))
// 		s.Authorities[index].SigningKey = *testHelper.NewPrimitivesPrivateKey(uint64(i + sigs)).Pub
// 		s.Authorities[index].Status = 0

// 		s.AddAuditServer(s.GetLeaderHeight(), testHelper.NewRepeatingHash(byte(i+sigs)))
// 	}

// 	err = sf.UpdateState(s)
// 	if err != nil {
// 		t.Errorf("%v", err)
// 	}

// }

/*
func TestAuthoritySignature(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	idindex := CreateBlankFactomIdentity(s, primitives.NewZeroHash())
	s.Identities[idindex].ManagementChainID = primitives.NewZeroHash()

	index := s.AddAuthorityFromChainID(primitives.NewZeroHash())
	s.Authorities[index].SigningKey = *(s.GetServerPublicKey())
	s.Authorities[index].Status = 1

	ack := new(messages.Ack)
	ack.DBHeight = s.LLeaderHeight
	ack.VMIndex = 1
	ack.Minute = byte(5)
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = primitives.NewZeroHash()
	ack.LeaderChainID = s.IdentityChainID
	ack.SerialHash = primitives.NewZeroHash()

	err := ack.Sign(s)
	if err != nil {
		t.Error("Authority Test Failed when signing message")
	}

	msg, err := ack.MarshalForSignature()
	if err != nil {
		t.Error("Authority Test Failed when marshalling for sig")
	}

	sig := ack.GetSignature()
	server, err := s.Authorities[0].VerifySignature(msg, sig.GetSignature())
	if !server || err != nil {
		t.Error("Authority Test Failed when checking sigs")
	}
}

*/
