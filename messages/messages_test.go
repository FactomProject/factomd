package messages

import (
	"testing"
)

func TestEomMessageReadString(t *testing.T) {
	var m EomMessage
	s := "EOM 1/2/3 ID-89abcdef"
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("EomMessage.ReadString(\"%s\")", s)
	}
}

func TestFaultMsgReadString(t *testing.T) {
	var m FaultMsg
	s := "FAULT ID-01234567 1/2/3 99 ID-89abcdef"
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("FaultMsg.ReadString(\"%s\")", s)
	}
}

func TestDbsigMessageReadString(t *testing.T) {
	var m DbsigMessage
	s := "DBSIG -000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f- 99 <EOM 1/2/3 ID-89abcdef> ID-89abcdef"
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("DbsigMessage.ReadString(\"%s\")", s)
	}
}

func TestAuthChangeMessageReadString(t *testing.T) {
	var m AuthChangeMessage
	s := "AUTH ID-76543210 LEADER ID-89abcdef"
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("AuthChangeMessage.ReadString(\"%s\")", s)
	}
	s = "AUTH ID-76543210 AUDIT ID-89abcedf"
	m.ReadString(s)
	r = m.String()
	if r != s {
		t.Errorf("AuthChangeMessage.ReadString(\"%s\")", s)
	}
}

func TestVolunteerMessageReadString(t *testing.T) {
	var m VolunteerMessage
	s := "VOLUNTEER ID-76543210 <EOM 1/2/3 ID-89abcdef> <FAULT ID-01234567 1/2/3 99 ID-89abcdef> ID-89abcdef"
	m.ReadString(s)
	r := m.String()
	if r != s {
		t.Errorf("VolunteerMessage.ReadString(\"%s\")", s)
	}
}