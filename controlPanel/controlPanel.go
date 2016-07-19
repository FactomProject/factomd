package controlPanel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/state"
)

var TEMPLATE_PATH string
var templates *template.Template

var INDEX_HTML []byte
var mux *http.ServeMux
var st *state.State
var index int = 0
var fnodes []*state.State

func ServeControlPanel(port int, states []*state.State) {
	defer func() {
		// recover from panic if files path is incorrect
		if recover() != nil {
			fmt.Println("Control Panel has encountered a panic and will not be served")
		}
	}()

	st = states[index]
	fnodes = states
	portStr := ":" + strconv.Itoa(port)

	//factomdDir := ""
	TEMPLATE_PATH = "./controlPanel/Web/templates/"
	templates = template.Must(template.ParseGlob(TEMPLATE_PATH + "general/*.html")) //Cache general templates

	fmt.Println("Starting Control Panel on http://localhost" + portStr + "/")
	RecentTransactions = new(LastDirectoryBlockTransactions)
	// Mux for static files
	mux = http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./controlPanel/Web")))

	INDEX_HTML, _ = ioutil.ReadFile("./controlPanel/Web/index.html")

	go doEvery(5*time.Second, getRecentTransactions)

	http.HandleFunc("/", static(indexHandler))
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/factomd", factomdHandler)

	http.ListenAndServe(portStr, nil)
}

func static(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.ContainsRune(r.URL.Path, '.') {
			mux.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ParseGlob(TEMPLATE_PATH + "/index/*.html")
	err := templates.ExecuteTemplate(w, "indexPage", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	method := r.FormValue("method")
	switch method {
	case "search":
		found, respose := searchDB(r.FormValue("search"), st)
		if found {
			w.Write([]byte(respose))
			return
		}
	}
	w.Write([]byte(`{"Type": "None"}`))
}

type SearchedStruct struct {
	Type    string      `json:"Type"`
	Content interface{} `json:"item"`

	Input string
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	searchResult := new(SearchedStruct)
	if r.Method == "POST" {
		data := r.FormValue("content")
		json.Unmarshal([]byte(data), searchResult)
	} else {
		searchResult.Type = r.FormValue("type")
	}
	searchResult.Input = r.FormValue("input")
	handleSearchResult(searchResult, w)
	//search, _ := ioutil.ReadFile("./ControlPanel/Web/searchresult.html")
	//w.Write([]byte(searchResult.Type))
}

func factomdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	item := r.FormValue("item") // Item wanted
	switch item {
	case "myHeight":
		data := fmt.Sprintf("%d", st.GetHighestRecordedBlock())
		w.Write([]byte(data)) // Return current node height
	case "leaderHeight":
		data := fmt.Sprintf("%d", st.GetLeaderHeight()-1)
		if st.GetLeaderHeight() == 0 {
			data = "0"
		}
		w.Write([]byte(data)) // Return leader height
	case "completeHeight": // Second Pass Sync info
		data := fmt.Sprintf("%d", st.GetEBDBHeightComplete())
		w.Write([]byte(data)) // Return EBDB complete height
	case "dataDump":
		data := getDataDumps()
		w.Write(data)
	case "nextNode":
		index++
		if index >= len(fnodes) {
			index = 0
		}
		w.Write([]byte(fmt.Sprintf("%d", index)))
	case "peers":
		data := getPeers()
		w.Write(data)
	case "recentTransactions":
		//data := getRecentTransactions()
		data, err := json.Marshal(RecentTransactions)
		if err != nil {
			data = []byte(`{"list":"none"}`)
		}
		w.Write(data)
	}
}

func getPeers() []byte {
	return []byte("")
}

type LastDirectoryBlockTransactions struct {
	DirectoryBlock struct {
		KeyMR     string
		BodyKeyMR string
		FullHash  string
		DBHeight  string

		PrevFullHash string
		PrevKeyMR    string
	}
	FactoidTransactions []struct {
		TxID         string
		TotalInput   string
		Status       string
		TotalInputs  int
		TotalOutputs int
	}
	Entries []EntryHolder
}

var RecentTransactions *LastDirectoryBlockTransactions

func doEvery(d time.Duration, f func(time.Time) []byte) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func getRecentTransactions(time.Time) []byte {
	last := st.GetDirectoryBlock()
	if last == nil {
		return []byte(`{"list":"none"}`)
	}

	if RecentTransactions == nil {
		return []byte(`{"list":"none"}`)
	}

	RecentTransactions.DirectoryBlock = struct {
		KeyMR     string
		BodyKeyMR string
		FullHash  string
		DBHeight  string

		PrevFullHash string
		PrevKeyMR    string
	}{last.GetKeyMR().String(), last.BodyKeyMR().String(), last.GetFullHash().String(), fmt.Sprintf("%d", last.GetDatabaseHeight()), last.GetHeader().GetPrevFullHash().String(), last.GetHeader().GetPrevKeyMR().String()}

	vms := st.LeaderPL.VMs
	for _, vm := range vms {
		if vm == nil {
			continue
		}
		for _, msg := range vm.List {
			if msg == nil {
				continue
			}
			switch msg.Type() {
			case constants.COMMIT_CHAIN_MSG:
			case constants.COMMIT_ENTRY_MSG:
			case constants.REVEAL_ENTRY_MSG:
				data, err := msg.MarshalBinary()
				if err != nil {
					continue
				}
				rev := new(messages.RevealEntryMsg)
				err = rev.UnmarshalBinary(data)
				if rev.Entry == nil || err != nil {
					continue
				}
				e := new(EntryHolder)
				ack := getEntryAck(rev.Entry.GetHash().String())
				if ack == nil {
					continue
				}
				e.Hash = ack.EntryHash
				e.ChainID = "Processing"
				has := false
				for _, ent := range RecentTransactions.Entries {
					if ent.Hash == e.Hash {
						has = true
						break
					}
				}
				if !has {
					RecentTransactions.Entries = append([]EntryHolder{*e}, RecentTransactions.Entries...)
				}
			case constants.FACTOID_TRANSACTION_MSG:
				data, err := msg.MarshalBinary()
				if err != nil {
					continue
				}
				transMsg := new(messages.FactoidTransaction)
				err = transMsg.UnmarshalBinary(data)
				if transMsg.Transaction == nil || err != nil {
					continue
				}
				trans := transMsg.Transaction
				input, err := trans.TotalInputs()
				if err != nil {
					continue
				}
				totalInputs := len(trans.GetInputs())
				totalOutputs := len(trans.GetECOutputs())
				totalOutputs = totalOutputs + len(trans.GetOutputs())
				inputStr := fmt.Sprintf("%f", float64(input)/1e8)
				has := false
				for _, fact := range RecentTransactions.FactoidTransactions {
					if fact.TxID == trans.GetHash().String() {
						has = true
						break
					}
				}
				if !has {
					RecentTransactions.FactoidTransactions = append([]struct {
						TxID         string
						TotalInput   string
						Status       string
						TotalInputs  int
						TotalOutputs int
					}{{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs}}, RecentTransactions.FactoidTransactions...)
				}
			}
		}
	}

	entries := last.GetDBEntries()
	//entries = append(entries, pl.DirectoryBlock.GetDBEntries()[:]...)
	for _, entry := range entries {
		if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000f" {
			mr := entry.GetKeyMR()
			fblock, err := st.DB.FetchFBlock(mr)
			if err != nil || fblock == nil {
				continue
			}
			transactions := fblock.GetTransactions()
			for _, trans := range transactions {
				input, err := trans.TotalInputs()
				if err != nil {
					continue
				}
				totalInputs := len(trans.GetInputs())
				totalOutputs := len(trans.GetECOutputs())
				totalOutputs = totalOutputs + len(trans.GetOutputs())
				inputStr := fmt.Sprintf("%f", float64(input)/1e8)
				has := false
				for i, fact := range RecentTransactions.FactoidTransactions {
					if fact.TxID == trans.GetHash().String() {
						//RecentTransactions.FactoidTransactions = append(RecentTransactions.FactoidTransactions[:i], RecentTransactions.FactoidTransactions[i+1:]...)
						RecentTransactions.FactoidTransactions[i] = struct {
							TxID         string
							TotalInput   string
							Status       string
							TotalInputs  int
							TotalOutputs int
						}{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs}
						has = true
						break
					}
				}
				if !has {
					RecentTransactions.FactoidTransactions = append([]struct {
						TxID         string
						TotalInput   string
						Status       string
						TotalInputs  int
						TotalOutputs int
					}{{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs}}, RecentTransactions.FactoidTransactions...)
				}
			}
		} else if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000c" {
			mr := entry.GetKeyMR()
			ecblock, err := st.DB.FetchECBlock(mr)
			if err != nil || ecblock == nil {
				continue
			}
			ents := ecblock.GetEntries()
			for _, entry := range ents {
				if entry.GetEntryHash() != nil {
					e := getEntry(entry.GetEntryHash().String())
					if e != nil {
						has := false
						for i, ent := range RecentTransactions.Entries {
							if ent.Hash == e.Hash {
								RecentTransactions.Entries[i] = *e
								has = true
								break
								//RecentTransactions.Entries = append(RecentTransactions.Entries[:i], RecentTransactions.Entries[i+1:]...)
							}
						}
						if !has {
							RecentTransactions.Entries = append([]EntryHolder{*e}, RecentTransactions.Entries...)
						}
					}
				}
			}
		}
	}

	if len(RecentTransactions.Entries) > 100 {
		RecentTransactions.Entries = RecentTransactions.Entries[:101]
	}
	if len(RecentTransactions.FactoidTransactions) > 100 {
		RecentTransactions.FactoidTransactions = RecentTransactions.FactoidTransactions[:101]
	}
	ret, err := json.Marshal(RecentTransactions)
	if err != nil {
		return []byte(`{"list":"none"}`)
	}
	return ret
}
