package controlPanel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
)

var (
	UpdateTimeValue int = 5 // in seconds. How long to update the state and recent transactions

	FILES_PATH string
	templates  *template.Template

	INDEX_HTML []byte
	mux        *http.ServeMux
	index      int = 0

	Fnodes       []*state.State
	StatePointer *state.State
	Controller   *p2p.Controller
	GitBuild     string

	// Sync Mutex
	TemplateMutex sync.Mutex
)

func directoryExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			return false
		}
	}
	return true
}

func ServeControlPanel(port int, states []*state.State, connections chan map[string]p2p.ConnectionMetrics, controller *p2p.Controller, gitBuild string) {
	defer func() {
		// recover from panic if files path is incorrect
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic.\n", r)
		}
	}()

	GitBuild = gitBuild
	portStr := ":" + strconv.Itoa(port)
	StatePointer = states[index]
	Fnodes = states
	Controller = controller

	// Load Files
	FILES_PATH = states[0].ControlPanelPath
	if !directoryExists(FILES_PATH) {
		FILES_PATH = "./controlPanel/Web/"
		if !directoryExists(FILES_PATH) {
			fmt.Println("Control Panel static files cannot be found.")
			http.HandleFunc("/", noStaticFilesFoundHandler)
			http.ListenAndServe(portStr, nil)
			return
		}
	}
	TemplateMutex.Lock()
	templates = template.Must(template.ParseGlob(FILES_PATH + "templates/general/*.html"))
	TemplateMutex.Unlock()

	// Updated Globals
	RecentTransactions = new(LastDirectoryBlockTransactions)
	AllConnections = new(ConnectionsMap)
	AllConnections.connected = map[string]p2p.ConnectionMetrics{}
	AllConnections.disconnected = map[string]p2p.ConnectionMetrics{}

	fmt.Println("Starting Control Panel on http://localhost" + portStr + "/")

	// Mux for static files
	mux = http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(FILES_PATH)))
	INDEX_HTML, _ = ioutil.ReadFile(FILES_PATH + "templates/index.html")

	go doEvery(5*time.Second, getRecentTransactions)
	go manageConnections(connections)

	http.HandleFunc("/", static(indexHandler))
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/factomd", factomdHandler)

	http.ListenAndServe(portStr, nil)
}

func noStaticFilesFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "The control panel was not able to be correctly loaded because the Web files were not found. "+
		"\nFactomd is looking in %s folder for the files, placing the \n"+
		"Web files in that directory should resolve this error.", Fnodes[0].ControlPanelPath)
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
	TemplateMutex.Lock()
	templates.ParseGlob(FILES_PATH + "templates/index/*.html")
	TemplateMutex.Unlock()
	if len(GitBuild) == 0 {
		GitBuild = "Unknown (Must install with script)"
	}
	err := templates.ExecuteTemplate(w, "indexPage", GitBuild)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic()
	if StatePointer.GetIdentityChainID() == nil {
		return
	}
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	method := r.FormValue("method")
	switch method {
	case "search":
		found, respose := searchDB(r.FormValue("search"), *StatePointer)
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
	defer recoverFromPanic()
	if StatePointer.GetIdentityChainID() == nil {
		return
	}
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
	defer recoverFromPanic()
	if StatePointer.GetIdentityChainID() == nil {
		return
	}
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	item := r.FormValue("item") // Item wanted
	switch item {
	case "myHeight":
		data := fmt.Sprintf("%d", StatePointer.GetHighestRecordedBlock())
		w.Write([]byte(data)) // Return current node height
	case "leaderHeight":
		data := fmt.Sprintf("%d", StatePointer.GetLeaderHeight()-1)
		if StatePointer.GetLeaderHeight() == 0 {
			data = "0"
		}
		w.Write([]byte(data)) // Return leader height
	case "completeHeight": // Second Pass Sync info
		data := fmt.Sprintf("%d", StatePointer.GetEBDBHeightComplete())
		w.Write([]byte(data)) // Return EBDB complete height
	case "connections":
	case "dataDump":
		data := getDataDumps()
		w.Write(data)
	case "nextNode":
		index++
		if index >= len(Fnodes) {
			index = 0
		}
		StatePointer = Fnodes[index]
		w.Write([]byte(fmt.Sprintf("%d", index)))
	case "peers":
		data := getPeers()
		w.Write(data)
	case "peerTotals":
		data := getPeetTotals()
		w.Write(data)
	case "recentTransactions":
		data := []byte(`{"list":"none"}`)
		var err error
		if RecentTransactions == nil {
			data = []byte(`{"list":"none"}`)
		} else {
			data, err = json.Marshal(RecentTransactions)
			if err != nil {
				data = []byte(`{"list":"none"}`)
			}
		}
		w.Write(data)
	case "disconnect":
		data := []byte(r.FormValue("value"))
		disconnectPeer(r.FormValue("value"))
		w.Write(data)
	}
}

func disconnectPeer(hash string) {
	if Controller != nil {
		fmt.Println("ControlPanel: Sent a disconnect signal.")
		Controller.Ban(hash)
	}
}

func getPeers() []byte {
	data, err := json.Marshal(AllConnections.SortedConnections())
	if err != nil {
		return []byte(`error`)
	}
	return data
}

func getPeetTotals() []byte {
	data, err := json.Marshal(AllConnections.totals)
	if err != nil {
		return []byte(`error`)
	}
	return data
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

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func getRecentTransactions(time.Time) {
	defer recoverFromPanic()
	if StatePointer == nil {
		return
	}
	last := StatePointer.GetDirectoryBlock()
	if last == nil {
		return
	}

	if RecentTransactions == nil {
		return
	}

	RecentTransactions.DirectoryBlock = struct {
		KeyMR     string
		BodyKeyMR string
		FullHash  string
		DBHeight  string

		PrevFullHash string
		PrevKeyMR    string
	}{last.GetKeyMR().String(), last.BodyKeyMR().String(), last.GetFullHash().String(), fmt.Sprintf("%d", last.GetDatabaseHeight()), last.GetHeader().GetPrevFullHash().String(), last.GetHeader().GetPrevKeyMR().String()}

	vms := StatePointer.LeaderPL.VMs
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
				/*ack := getEntryAck(rev.Entry.GetHash().String())
				if ack == nil {
					continue
				}*/
				e.Hash = rev.Entry.GetHash().String()
				e.ChainID = "Processing"
				has := false
				for _, ent := range RecentTransactions.Entries {
					if ent.Hash == e.Hash {
						has = true
						break
					}
				}
				if !has {
					RecentTransactions.Entries = append(RecentTransactions.Entries, *e)
					//RecentTransactions.Entries = append([]EntryHolder{*e}, RecentTransactions.Entries...)
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
					RecentTransactions.FactoidTransactions = append(RecentTransactions.FactoidTransactions, struct {
						TxID         string
						TotalInput   string
						Status       string
						TotalInputs  int
						TotalOutputs int
					}{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs})
					/*RecentTransactions.FactoidTransactions = append([]struct {
						TxID         string
						TotalInput   string
						Status       string
						TotalInputs  int
						TotalOutputs int
					}{{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs}}, RecentTransactions.FactoidTransactions...)
					*/
				}
			}
		}
	}

	entries := last.GetDBEntries()
	//entries = append(entries, pl.DirectoryBlock.GetDBEntries()[:]...)
	for _, entry := range entries {
		if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000f" {
			mr := entry.GetKeyMR()
			fblock, err := StatePointer.DB.FetchFBlock(mr)
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
						RecentTransactions.FactoidTransactions[i] = struct {
							TxID         string
							TotalInput   string
							Status       string
							TotalInputs  int
							TotalOutputs int
						}{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs}
						//RecentTransactions.FactoidTransactions = append(RecentTransactions.FactoidTransactions[:i], RecentTransactions.FactoidTransactions[i+1:]...)
						/*RecentTransactions.FactoidTransactions[i] = struct {
							TxID         string
							TotalInput   string
							Status       string
							TotalInputs  int
							TotalOutputs int
						}{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs}*/
						has = true
						break
					}
				}
				if !has {
					RecentTransactions.FactoidTransactions = append(RecentTransactions.FactoidTransactions, struct {
						TxID         string
						TotalInput   string
						Status       string
						TotalInputs  int
						TotalOutputs int
					}{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs})
					/*RecentTransactions.FactoidTransactions = append([]struct {
						TxID         string
						TotalInput   string
						Status       string
						TotalInputs  int
						TotalOutputs int
					}{{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs}}, RecentTransactions.FactoidTransactions...)*/
				}
			}
		} else if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000c" {
			mr := entry.GetKeyMR()
			ecblock, err := StatePointer.DB.FetchECBlock(mr)
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
							RecentTransactions.Entries = append(RecentTransactions.Entries, *e)
							//RecentTransactions.Entries = append([]EntryHolder{*e}, RecentTransactions.Entries...)
						}
					}
				}
			}
		}
	}

	if len(RecentTransactions.Entries) > 100 {
		overflow := len(RecentTransactions.Entries) - 100
		RecentTransactions.Entries = RecentTransactions.Entries[overflow:]
	}
	if len(RecentTransactions.FactoidTransactions) > 100 {
		overflow := len(RecentTransactions.FactoidTransactions) - 100
		RecentTransactions.FactoidTransactions = RecentTransactions.FactoidTransactions[overflow:]
	}
	//_, err := json.Marshal(RecentTransactions)
	//if err != nil {
	//	return
	//}
	//return ret
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		fmt.Println("ERROR: Control Panel has encountered a panic and was halted. Reloading...\n", r)
	}
}
