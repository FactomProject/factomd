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

	//"github.com/FactomProject/factomd/common/constants"
	//"github.com/FactomProject/factomd/common/messages"
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

	DisplayState state.DisplayState
	StatePointer *state.State
	Controller   *p2p.Controller
	GitBuild     string

	LastRequest     time.Time
	TimeRequestHold float64 = 3 // Amount of time in seconds before can request data again

	DisplayStateChannel chan state.DisplayState

	// Sync Mutex
	TemplateMutex     sync.Mutex
	DisplayStateMutex sync.RWMutex
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

func DisplayStateDrain(channel chan state.DisplayState) {
	for {
		select {
		case ds := <-channel:
			DisplayStateMutex.Lock()
			DisplayState = ds
			DisplayStateMutex.Unlock()
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}
}

func ServeControlPanel(displayStateChannel chan state.DisplayState, statePointer *state.State, connections chan interface{}, controller *p2p.Controller, gitBuild string) {
	StatePointer = statePointer
	StatePointer.ControlPanelDataRequest = true
	// Wait for initial State
	select {
	case DisplayState = <-displayStateChannel:
		fmt.Println("Found state, control panel now active")
	}

	DisplayStateMutex.RLock()
	controlPanelSetting := DisplayState.ControlPanelSetting
	port := DisplayState.ControlPanelPort
	FILES_PATH = DisplayState.ControlPanelPath
	DisplayStateMutex.RUnlock()

	if controlPanelSetting == 0 {
		fmt.Println("Control Panel has been disabled withing the config file and will not be served. This is reccomened for any public server, if you wish to renable it, check your config file.")
		return
	}

	go DisplayStateDrain(displayStateChannel)

	GitBuild = gitBuild
	portStr := ":" + strconv.Itoa(port)
	Controller = controller

	// Load Files
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
	AllConnections = NewConnectionsMap()

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
	http.HandleFunc("/factomdBatch", factomdBatchHandler)

	http.ListenAndServe(portStr, nil)
}

func noStaticFilesFoundHandler(w http.ResponseWriter, r *http.Request) {
	DisplayStateMutex.RLock()
	Path := DisplayState.ControlPanelPath
	DisplayStateMutex.RUnlock()
	fmt.Fprintf(w, "The control panel was not able to be correctly loaded because the Web files were not found. "+
		"\nFactomd is looking in %s folder for the files, placing the \n"+
		"Web files in that directory should resolve this error.", Path)
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
	//defer recoverFromPanic()
	defer func() {
		// recover from panic if files path is incorrect
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic.\n", r)
		}
	}()
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
	//defer recoverFromPanic()
	defer func() {
		// recover from panic if files path is incorrect
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic.\n", r)
		}
	}()
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

var batchQueried = false

// Batches Json in []byte form to an array of json []byte objects
func factomdBatchHandler(w http.ResponseWriter, r *http.Request) {
	//defer recoverFromPanic()
	requestData()
	batchQueried = true
	if r.Method != "GET" {
		return
	}
	batch := r.FormValue("batch")
	batchData := make([]byte, 0)
	batchData = append(batchData, []byte(`[`)...)

	items := strings.Split(batch, ",")
	for _, item := range items {
		data := factomdQuery(item, "")
		batchData = append(batchData, data...)
		batchData = append(batchData, []byte(`,`)...)
	}

	batchQueried = false

	batchData = batchData[:len(batchData)-1]
	batchData = append(batchData, []byte(`]`)...)
	w.Write(batchData)
}

func factomdHandler(w http.ResponseWriter, r *http.Request) {
	//defer recoverFromPanic()
	defer func() {
		// recover from panic if files path is incorrect
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic.\n", r)
		}
	}()
	if r.Method != "GET" {
		//http.NotFound(w, r)
		return
	}
	item := r.FormValue("item")   // Item wanted
	value := r.FormValue("value") // Optional argument
	data := factomdQuery(item, value)
	w.Write([]byte(data))
}

// Flag to tell if data is already being requested
var requestMutex bool = false

func requestData() {
	if requestMutex {
		return
	}
	requestMutex = true
	if (time.Since(LastRequest)).Seconds() < TimeRequestHold {
		requestMutex = false
		return
	}
	LastRequest = time.Now()
	StatePointer.ControlPanelDataRequest = true
	requestMutex = false
}

func factomdQuery(item string, value string) []byte {
	if !batchQueried {
		requestData()
	}
	switch item {
	case "myHeight":
		DisplayStateMutex.RLock()
		h := DisplayState.CurrentNodeHeight
		DisplayStateMutex.RUnlock()
		return HeightToJsonStruct(h)
	case "leaderHeight":
		DisplayStateMutex.RLock()
		h := DisplayState.CurrentLeaderHeight - 1
		DisplayStateMutex.RUnlock()
		return HeightToJsonStruct(h)
	case "completeHeight": // Second Pass Sync info
		DisplayStateMutex.RLock()
		h := DisplayState.CurrentEBDBHeight
		DisplayStateMutex.RUnlock()
		return HeightToJsonStruct(h)
	case "connections":
	case "dataDump":
		data := getDataDumps()
		return data
	case "nextNode":
		index := 0
		/*index++
		if index >= len(Fnodes) {
			index = 0
		}
		DisplayState = Fnodes[index]*/
		return []byte(fmt.Sprintf("%d", index))
	case "channelLength":
		return []byte(fmt.Sprintf(`{"length":%d}`, len(DisplayStateChannel)))
	case "peers":
		data := getPeers()
		return data
	case "peerTotals":
		data := getPeetTotals()
		return data
	case "recentTransactions":
		RecentTransactionsMutex.Lock()
		defer RecentTransactionsMutex.Unlock()
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
		return data
	case "disconnect":
		hash := ""
		if len(value) > 0 {
			hash = hashPeerAddress(value)
		}
		DisplayStateMutex.RLock()
		CPS := DisplayState.ControlPanelSetting
		DisplayStateMutex.RUnlock()
		if CPS == 2 {
			disconnectPeer(value)
			return []byte(`{"Access":"granted", "Id":"` + hash + `"}`)
		} else {
			return []byte(`{"Access":"denied", "Id":"` + hash + `"}`)
		}
	}
	return []byte("")
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
	AllConnections.Lock.Lock()
	data, err := json.Marshal(AllConnections.Totals)
	AllConnections.Lock.Unlock()
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

// Flag to tell if RecentTransactions is already being built
var DoingRecentTransactions bool
var RecentTransactionsMutex sync.Mutex

func toggleDCT() {
	if DoingRecentTransactions {
		DoingRecentTransactions = false
	} else {
		DoingRecentTransactions = true
	}
}

func getRecentTransactions(time.Time) {
	if DoingRecentTransactions {
		return
	}
	toggleDCT()
	defer toggleDCT()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic.\n", r)
		}
	}()
	if StatePointer == nil {
		return
	}
	DisplayStateMutex.RLock()
	last := DisplayState.LastDirectoryBlock
	DisplayStateMutex.RUnlock()
	if last == nil {
		return
	}

	RecentTransactionsMutex.Lock()
	defer RecentTransactionsMutex.Unlock()

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

	for _, entry := range DisplayState.PLEntry {
		e := new(EntryHolder)
		e.Hash = entry.EntryHash
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
		}
	}

	for _, fTrans := range DisplayState.PLFactoid {
		has := false
		for _, trans := range RecentTransactions.FactoidTransactions {
			if fTrans.TxID == trans.TxID {
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
			}{fTrans.TxID, fTrans.TotalInput, "Processing", fTrans.TotalInputs, fTrans.TotalOutputs})
		}
	}

	entries := last.GetDBEntries()
	for _, entry := range entries {
		if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000f" {
			mr := entry.GetKeyMR()
			dbase := StatePointer.GetAndLockDB()
			fblock, err := dbase.FetchFBlock(mr)
			if err != nil || fblock == nil {
				StatePointer.UnlockDB()
				continue
			}
			StatePointer.UnlockDB()
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
				}
			}
		} else if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000c" {
			mr := entry.GetKeyMR()

			dbase := StatePointer.GetAndLockDB()
			ecblock, err := dbase.FetchECBlock(mr)
			StatePointer.UnlockDB()
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
							}
						}
						if !has {
							RecentTransactions.Entries = append(RecentTransactions.Entries, *e)
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
}

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}
