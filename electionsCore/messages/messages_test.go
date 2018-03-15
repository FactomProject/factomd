package messages

import (
	. "github.com/FactomProject/factomd/electionsCore/errorhandling"
	. "github.com/FactomProject/factomd/electionsCore/primitives"
	"testing"
)

// string to identity
func s2i(s string) Identity {
	var i Identity
	i.ReadString(s)
	return i
}

// string to process List Location
func s2pl(s string) ProcessListLocation {
	var i ProcessListLocation
	i.ReadString(s)
	return i
}

func TestEomMessageReadString(t *testing.T) {
	T = t // Set test for error handling
	var m EomMessage
	s := `EomMessage {"Vm":1,"Minute":2,"Height":3,"Signer":"ID-89abcdef"}`
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("EomMessage.ReadString(\"%s\")", s)
	}
}

func TestFaultMsgReadString(t *testing.T) {
	T = t // Set test for error handling
	var m FaultMsg
	s := `FaultMsg {"FaultId":"ID-00000001","Vm":4,"Minute":3,"Height":2,"Round":1,"Signer":"ID-deadbeef"}`
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("FaultMsg.ReadString(\"%s\")", s)
	}
}

func TestDbsigMessageReadString(t *testing.T) {
	T = t // Set test for error handling
	var m DbsigMessage
	s := `DbsigMessage {"Prev":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"Height":0,"Eom":{"Vm":0,"Minute":0,"Height":0,"Signer":"ID-00000000"},"Signer":"ID-00000000"}`
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("DbsigMessage.ReadString(\"%s\")", s)
	}
}

func TestAuthChangeMessageReadString(t *testing.T) {
	T = t // Set test for error handling
	var m AuthChangeMessage

	s := `AuthChangeMessage {"Id":"ID-01234567","Status":"LEADER","Signer":"ID-89abcdef"}`
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("AuthChangeMessage.ReadString(\"%s\")", s)
	}
	s = `AuthChangeMessage {"Id":"ID-01234567","Status":"AUDIT","Signer":"ID-89abcdef"}`
	m.ReadString(s)
	r = m.String()
	if r != s {
		t.Errorf("AuthChangeMessage.ReadString(\"%s\")", s)
	}
}

func TestVolunteerMessageReadString(t *testing.T) {
	T = t // Set test for error handling
	var m VolunteerMessage

	s := `VolunteerMessage {"Id":"ID-12345678","Eom":{"Vm":99,"Minute":98,"Height":97,"Signer":"ID-87654321"},"FaultId":"ID-deadbeef","Vm":96,"Minute":95,"Height":94,"Round":93,"Signer":"ID-12344321"}`
	m.ReadString(s)
	r := m.String()
	//	fmt.Printf("s:%s:%d\nr:%s:%d\n",s,len(s),r,len(r))
	if r != s {
		t.Errorf("VolunteerMessage.ReadString(\"%s\")", s)
	}
}

func TestVoteMessageReadString(t *testing.T) {
	T = t // Set test for error handling
	var m VoteMessage
	x := m.String()
	_ = x

	s := `VoteMessages {"Volunteer":{"Id":"ID-00000000","Eom":{"Vm":0,"Minute":0,"Height":0,"Signer":"ID-00000000"},"FaultId":"ID-00000000","Vm":0,"Minute":0,"Height":0,"Round":0,"Signer":"ID-00000000"},"OtherVotes":null,"Signer":"ID-00000000"}`
	m.ReadString(s)
	r := m.String()
	//	fmt.Printf("s:%s:%d\nr:%s:%d\n",s,len(s),r,len(r))
	if r != s {
		t.Errorf("VoteMessages.ReadString(\"%s\")", s)
	}
}
