package controlPanel_test

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	//"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	//"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/testHelper"
)

var _ = time.Now()

var _ = fmt.Sprintf("")

// Enable for long test
var LongTest bool = false

func TestFactoidHas(t *testing.T) {
	rc := new(LastDirectoryBlockTransactions)
	rc.FactoidTransactions = append(rc.FactoidTransactions)
	for i := 0; i < 10; i++ {
		addTrans(rc)
	}

	for i := 0; i < len(rc.FactoidTransactions); i++ {
		if !rc.ContainsTrans(rc.FactoidTransactions[i].TxID) {
			t.Error("This should be true")
		}
	}

	for i := 0; i < len(rc.Entries); i++ {
		if !rc.ContainsEntry(rc.Entries[i].Hash) {
			t.Error("This should be true")
		}
	}
}

func addTrans(rc *LastDirectoryBlockTransactions) {
	rc.FactoidTransactions = append(rc.FactoidTransactions, struct {
		TxID         string
		Hash         string
		TotalInput   string
		Status       string
		TotalInputs  int
		TotalOutputs int
	}{primitives.RandomHash().String(), primitives.RandomHash().String(), "1", "Confirmed", 1, 1})

	e := new(EntryHolder)
	e.Hash = primitives.RandomHash().String()
	rc.Entries = append(rc.Entries, *e)
}

func TestControlPanel(t *testing.T) {
	if LongTest {
		var i uint32
		connections := make(chan interface{})
		emptyState := CreateAndPopulateTestStateAndStartValidator()

		gitBuild := "Test Is Running"
		go ServeControlPanel(emptyState.ControlPanelChannel, emptyState, connections, nil, gitBuild, "Fnode0")
		emptyState.CopyStateToControlPanel()
		for count := 0; count < 1000; count++ {
			for i = 0; i < 5; i++ {
				PopulateConnectionChan(i, connections)

			}
			for i = 5; i > 0; i-- {
				PopulateConnectionChan(i, connections)
			}
		}
	}
}

func TestDataDump(t *testing.T) {
	AllConnections = NewConnectionsMap()
	s := CreateAndPopulateTestStateAndStartValidator()
	ds, err := state.DeepStateDisplayCopy(s)
	if err != nil {
		t.Error(err)
	}

	DisplayState = *ds
	d := GetDataDumps()
	if len(d) == 0 {
		t.Error("No data")
	}
}

func TestSearching(t *testing.T) {
	var err error
	InitTemplates()
	s := CreateAndPopulateTestStateAndStartValidator()
	StatePointer = s

	c := new(SearchedStruct)
	c.Type = "entry"
	c.Input = ""

	// Search for an entry
	var e interfaces.IEBEntry
	i := uint32(0)
	for e == nil {
		d, err := s.DB.FetchDBlockByHeight(i)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if len(d.GetEBlockDBEntries()) > 0 {
			for _, ebhash := range d.GetEBlockDBEntries() {
				eblock, err := s.DB.FetchEBlock(ebhash.GetKeyMR())
				if err != nil {
					t.Error(err)
					t.FailNow()
				}

				for _, ehash := range eblock.GetEntryHashes() {
					if ehash.IsMinuteMarker() == true {
						continue
					}
					e, err = s.DB.FetchEntry(ehash)
					if err != nil {
						t.Error(err)
						t.FailNow()
					}

					c.Input = e.GetHash().String()
					content, err := searchfor(c)
					if err != nil {
						t.Error(err)
						t.FailNow()
					}

					if !strings.Contains(string(content), e.GetChainID().String()) {
						t.Error("Does not contain correct content")
					}
				}
			}
		}
	}

	// Insert a special crafted entry
	entry := entryBlock.NewEntry()
	var one primitives.ByteSlice
	one.Bytes = []byte{0xFF, 0x00}
	entry.ExtIDs = []primitives.ByteSlice{one}
	entry.GetHash()

	db := s.DB.(*databaseOverlay.Overlay)
	db.InsertEntry(entry)

	c.Input = entry.GetHash().String()
	content, err := searchfor(c)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !strings.Contains(string(content), entry.GetChainID().String()) {
		t.Error("Does not contain correct content")
	}

	c.Input = primitives.RandomHash().String()
	content, err = searchfor(c)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !strings.Contains(string(content), "It seems there has been an error while ") {
		t.Error("Does not contain correct content")
	}
}

// Trip some code that is sometimes tested in other unit tests. These lines don't have
func TestRequest(t *testing.T) {
	LastRequest = time.Time{}

	SetRequestMutex(true)
	RequestData()
	if time.Since(LastRequest).Seconds() < 1 {
		t.Error("Should have not updated the time")
	}

	SetRequestMutex(false)
	RequestData()
	if time.Since(LastRequest).Seconds() > 1 {
		t.Error("Should have updated the time")
	}

	RequestData()
}

func searchfor(ss *SearchedStruct) ([]byte, error) {
	w := httptest.NewRecorder()
	HandleSearchResult(ss, w)
	return ioutil.ReadAll(w.Body)
}
