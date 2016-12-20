package engine_test

import (
	"testing"
	"time"

	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestLoadJournalFromReader(t *testing.T) {
	journalStr := CreateTestLogFileString()
	s := CreateEmptyTestState()
	go s.ValidatorLoop()

	LoadJournalFromString(s, journalStr)
	time.Sleep(time.Second)

	head := s.GetDBHeightComplete()

	if head != 9 {
		t.Errorf("Head is %v, expected 9", head)
	}
}
