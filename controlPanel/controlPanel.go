package controlPanel

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel/files"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"

	llog "github.com/FactomProject/factomd/log"
)

// Initiates control panel variables and controls the http requests

type IndexTemplateData struct {
	GitBuild string
	Version  string
	NodeName string
}

var (
	UpdateTimeValue int = 5 // in seconds. How long to update the state and recent transactions

	//FILES_PATH string
	templates *template.Template

	//INDEX_HTML []byte
	mux   *http.ServeMux
	index int = 0

	DisplayState      state.DisplayState
	StatePointer      *state.State
	Controller        *p2p.Network // Used for Disconnect
	indexTemplateData *IndexTemplateData

	LastRequest     time.Time
	TimeRequestHold float64 = 3 // Amount of time in seconds before can request data again

	DisplayStateChannel chan state.DisplayState

	// Sync Mutex
	TemplateMutex     sync.Mutex
	DisplayStateMutex sync.RWMutex

	// This is the cached dblock for the getRecentTransactions thread
	getRecentTransactionsDBlock interfaces.IDirectoryBlock
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
			// Keep our block if there is not a new dblock
			dblock := DisplayState.LastDirectoryBlock
			DisplayState = ds
			if ds.LastDirectoryBlock == nil {
				DisplayState.LastDirectoryBlock = dblock
			}
			DisplayStateMutex.Unlock()
		default:
			RequestData()
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func InitTemplates() {
	TemplateMutex.Lock()
	templates = files.CustomParseGlob(nil, "templates/general/*.html")
	templates = template.Must(templates, nil)
	TemplateMutex.Unlock()
}

// Main function. This intiates appropriate variables and starts the control panel serving
func ServeControlPanel(displayStateChannel chan state.DisplayState, statePointer *state.State, connections chan map[string]p2p.PeerMetrics, controller *p2p.Network, gitBuild string, nodeName string) {
	defer func() {
		if r := recover(); r != nil {
			// The following recover string indicates an overwrite of existing http.ListenAndServe goroutine
			if r != "http: multiple registrations for /" {
				fmt.Println("Control Panel has encountered a panic in ServeControlPanel.\n", r)
			}
			llog.LogPrintf("recovery", "Control Panel has encountered a panic in ServeControlPanel. %v", r)
		}
	}()

	StatePointer = statePointer
	StatePointer.ControlPanelDataRequest = true // Request initial State
	// Wait for initial State
	DisplayState = <-displayStateChannel

	DisplayStateMutex.RLock()
	controlPanelSetting := DisplayState.ControlPanelSetting
	port := DisplayState.ControlPanelPort
	DisplayStateMutex.RUnlock()

	if controlPanelSetting == 0 { // 0 = Disabled
		fmt.Println("Control Panel has been disabled within the config file and will not be served. This is recommended for any public server, if you wish to renable it, check your config file.")
		return
	}

	go DisplayStateDrain(displayStateChannel)

	indexTemplateData = new(IndexTemplateData)
	indexTemplateData.GitBuild = gitBuild
	indexTemplateData.NodeName = nodeName
	indexTemplateData.Version = statePointer.GetFactomdVersion()
	portStr := ":" + strconv.Itoa(port)
	Controller = controller
	InitTemplates()

	// Updated Globals. A separate GoRoutine updates these, we just initialize
	RecentTransactions = new(LastDirectoryBlockTransactions)
	AllConnections = NewConnectionsMap()

	// Mux for static files
	mux = http.NewServeMux()
	mux.Handle("/", files.StaticServer)

	go doEvery(10*time.Second, getRecentTransactions)
	go manageConnections(connections)

	controlPanelMux := http.NewServeMux()
	controlPanelMux.HandleFunc("/", static(indexHandler))
	controlPanelMux.HandleFunc("/search", searchHandler)
	controlPanelMux.HandleFunc("/post", postHandler)
	controlPanelMux.HandleFunc("/factomd", factomdHandler)
	controlPanelMux.HandleFunc("/factomdBatch", factomdBatchHandler)

	tlsIsEnabled, tlsPrivate, tlsPublic := StatePointer.GetTlsInfo()
	if tlsIsEnabled {
	waitfortls:
		for {
			// lets wait for both the tls cert and key to be created.  if they are not created, wait for the RPC API process to create the files.
			// it is in a different goroutine, so just wait until it is done.  it happens in wsapi.Start with genCertPair()
			if _, err := os.Stat(tlsPublic); err == nil {
				if _, err := os.Stat(tlsPrivate); err == nil {
					break waitfortls
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Println("Starting encrypted Control Panel on https://localhost" + portStr + "/  Please note the HTTPS in the browser.")
		http.ListenAndServeTLS(portStr, tlsPublic, tlsPrivate, controlPanelMux)
	} else {
		fmt.Println("Starting Control Panel on http://localhost" + portStr + "/")
		http.ListenAndServe(portStr, controlPanelMux)
	}
}

func noStaticFilesFoundHandler(w http.ResponseWriter, r *http.Request) {
	DisplayStateMutex.RLock()
	DisplayStateMutex.RUnlock()
	fmt.Fprintf(w, "The control panel was not able to be correctly loaded because the Web files were not found. \n")
}

// For all static files. (CSS, JS, IMG, etc...)
func static(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if false == checkControlPanelPassword(w, r) {
			return
		}
		if strings.ContainsRune(r.URL.Path, '.') {
			mux.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic in IndexHandler.\n", r)
			llog.LogPrintf("recovery", "Control Panel has encountered a panic in IndexHandler. %v", r)
		}
	}()
	TemplateMutex.Lock()
	defer TemplateMutex.Unlock()
	if false == checkControlPanelPassword(w, r) {
		return
	}
	//templates.ParseGlob(FILES_PATH + "templates/index/*.html")
	files.CustomParseGlob(templates, "templates/index/*.html")
	if len(indexTemplateData.GitBuild) == 0 {
		indexTemplateData.GitBuild = "Unknown (Must install with script)"
	}

	err := templates.ExecuteTemplate(w, "indexPage", indexTemplateData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic in PostHandler.\n", r)
			llog.LogPrintf("recovery", "Control Panel has encountered a panic in PostHandler. %v", r)
		}
	}()
	if false == checkControlPanelPassword(w, r) {
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
		} else {
			if r.FormValue("known") == "factoidack" {
				w.Write([]byte(`{"Type": "special-action-fack"}`))
				return
			}
		}
	case "changelogs":
		// >= 2 means we have write access
		if StatePointer.ControlPanelSetting == 2 {
			newRegex := r.FormValue("logsetting")
			fmt.Printf("Changing log regex to: '%s'\n", newRegex)
			globals.Params.DebugLogRegEx = newRegex
		} else {
			w.Write([]byte(`{"Error": "Access denied"}`))
			return
		}
	}
	w.Write([]byte(`{"Type": "None"}`))
}

type SearchedStruct struct {
	Type    string      `json:"type"`
	Content interface{} `json:"item"`

	Input string `json:"input"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic in SearchHandler.\n", r)
			llog.LogPrintf("recovery", "Control Panel has encountered a panic in SearchHandler. %v", r)
		}
	}()
	if false == checkControlPanelPassword(w, r) {
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
	HandleSearchResult(searchResult, w)
}

// Batches Json in []byte form to an array of json []byte objects
func factomdBatchHandler(w http.ResponseWriter, r *http.Request) {
	if false == checkControlPanelPassword(w, r) {
		return
	}
	RequestData()
	if r.Method != "GET" {
		return
	}
	batch := r.FormValue("batch")
	batchData := make([]byte, 0)
	batchData = append(batchData, []byte(`[`)...)

	items := strings.Split(batch, ",")
	for _, item := range items {
		data := factomdQuery(item, "", true)
		batchData = append(batchData, data...)
		batchData = append(batchData, []byte(`,`)...)
	}

	batchData = batchData[:len(batchData)-1]
	batchData = append(batchData, []byte(`]`)...)
	w.Write(batchData)
}

func factomdHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic in FactomdHandler.\n", r)
			llog.LogPrintf("recovery", "Control Panel has encountered a panic in FactomdHandler. %v", r)
		}
	}()
	if false == checkControlPanelPassword(w, r) {
		return
	}
	if r.Method != "GET" {
		return
	}
	item := r.FormValue("item")   // Item wanted
	value := r.FormValue("value") // Optional argument
	data := factomdQuery(item, value, false)
	w.Write([]byte(data))
}

// Flag to tell if data is already being requested
var requestMutex bool = false

func SetRequestMutex(set bool) {
	requestMutex = set
}

func RequestData() {
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

func factomdQuery(item string, value string, batchQueried bool) []byte {
	if !batchQueried {
		RequestData()
	}
	switch item {
	case "ignoreDone":
		DisplayStateMutex.RLock()
		flag := DisplayState.IgnoreDone
		DisplayStateMutex.RUnlock()

		if flag {
			return []byte(`{"IgnoreDone": true}`)
		} else {
			return []byte(`{"IgnoreDone": false}`)
		}
	case "myHeight":
		DisplayStateMutex.RLock()
		h := DisplayState.CurrentNodeHeight
		DisplayStateMutex.RUnlock()
		return HeightToJsonStruct(h)
	case "leaderHeight":
		DisplayStateMutex.RLock()
		h := DisplayState.LeaderHeight
		if DisplayState.CurrentNodeHeight > DisplayState.LeaderHeight {
			h = DisplayState.CurrentNodeHeight
		}
		DisplayStateMutex.RUnlock()
		return HeightToJsonStruct(h)
	case "completeHeight": // Second Pass Sync info
		DisplayStateMutex.RLock()
		h := DisplayState.CurrentEBDBHeight
		DisplayStateMutex.RUnlock()
		return HeightToJsonStruct(h)
	case "connections":
	case "dataDump":
		data := GetDataDumps()
		return data
	case "nextNode":
		// Disabled
		index := 0
		/*index++
		if index >= len(Fnodes) {
			index = 0
		}
		DisplayState = Fnodes[index]*/
		return []byte(fmt.Sprintf("%d", index))
	case "servercount": // TODO
		DisplayStateMutex.RLock()
		feds := 0
		auds := 0
		for _, a := range DisplayState.Authorities {
			if a.Status == 1 {
				feds++
			} else if a.Status == 2 {
				auds++
			}
		}
		DisplayStateMutex.RUnlock()
		return []byte(fmt.Sprintf(`{"fed":%d,"aud":%d}`, feds, auds))
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
		Controller.Disconnect(hash)
	}
}

func getPeers() []byte {
	data, err := json.Marshal(AllConnections.SortedConnections())
	if err != nil {
		return []byte(`error`)
	}
	return data
}

// Returns the total and average statistics for the peer table
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
		Timestamp string

		PrevFullHash string
		PrevKeyMR    string
	}
	FactoidTransactions []struct {
		TxID         string
		Hash         string
		TotalInput   string
		Status       string
		TotalInputs  int
		TotalOutputs int
	}
	Entries []EntryHolder

	LastHeightChecked uint32
}

func (d *LastDirectoryBlockTransactions) ContainsEntry(hash string) bool {
	for _, entry := range d.Entries {
		if entry.Hash == hash {
			return true
		}
	}
	return false
}

func (d *LastDirectoryBlockTransactions) ContainsTrans(txid string) bool {
	for _, trans := range d.FactoidTransactions {
		if trans.TxID == txid {
			return true
		}
	}
	return false
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

// Gets all the recent transitions. Will only keep the most recent 100.
func getRecentTransactions(time.Time) {
	/*defer func() {
		if r := recover(); r != nil {
			fmt.Println("Control Panel has encountered a panic in GetRecentTransactions.\n", r)
		}
	}()*/

	if DoingRecentTransactions {
		return
	}
	toggleDCT()
	defer toggleDCT()

	if StatePointer == nil {
		return
	}

	var last interfaces.IDirectoryBlock
	// Need to copy the dblock as this runs on it's own thread. Only copy if it is new
	DisplayStateMutex.RLock()
	if DisplayState.LastDirectoryBlock == nil {
		DisplayStateMutex.RUnlock()
		return
	}

	// If our cached dblock is the same as the new one, don't make another copy
	if getRecentTransactionsDBlock == nil || DisplayState.LastDirectoryBlock.GetDatabaseHeight() != getRecentTransactionsDBlock.GetDatabaseHeight() {
		data, err := DisplayState.LastDirectoryBlock.MarshalBinary()
		DisplayStateMutex.RUnlock()
		if err != nil {
			return
		}

		last, err = directoryBlock.UnmarshalDBlock(data)
		if err != nil {
			return
		}
		getRecentTransactionsDBlock = last
	} else {
		DisplayStateMutex.RUnlock()
		last = getRecentTransactionsDBlock
	}

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
		Timestamp string

		PrevFullHash string
		PrevKeyMR    string
	}{last.GetKeyMR().String(), last.BodyKeyMR().String(), last.GetFullHash().String(), fmt.Sprintf("%d", last.GetDatabaseHeight()), last.GetTimestamp().String(), last.GetHeader().GetPrevFullHash().String(), last.GetHeader().GetPrevKeyMR().String()}
	// Process list items
	DisplayStateMutex.RLock()
	for i, entry := range DisplayState.PLEntry {
		e := new(EntryHolder)
		e.Hash = entry.EntryHash
		e.ChainID = "Processing"
		if !RecentTransactions.ContainsEntry(e.Hash) {
			RecentTransactions.Entries = append(RecentTransactions.Entries, *e)
		}
		if i > 100 {
			break // We only care about top 100
		}
	}

	for i, fTrans := range DisplayState.PLFactoid {
		if fTrans.TotalInputs == 0 {
			continue
		}
		if !RecentTransactions.ContainsTrans(fTrans.TxID) {
			RecentTransactions.FactoidTransactions = append(RecentTransactions.FactoidTransactions, struct {
				TxID         string
				Hash         string
				TotalInput   string
				Status       string
				TotalInputs  int
				TotalOutputs int
			}{fTrans.TxID, fTrans.Hash, fTrans.TotalInput, "Processing", fTrans.TotalInputs, fTrans.TotalOutputs})
		}
		if i > 100 {
			break // We only care about top 100
		}
	}
	DisplayStateMutex.RUnlock()

	entries := last.GetDBEntries()
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000f" {
			mr := entry.GetKeyMR()
			dbase := StatePointer.GetDB()
			fblock, err := dbase.FetchFBlock(mr)

			if err != nil || fblock == nil {
				continue
			}
			transactions := fblock.GetTransactions()
			for _, trans := range transactions {
				input, err := trans.TotalInputs()
				if err != nil || input == 0 {
					continue
				}
				totalInputs := len(trans.GetInputs())
				totalOutputs := len(trans.GetECOutputs())
				totalOutputs = totalOutputs + len(trans.GetOutputs())
				inputStr := fmt.Sprintf("%f", float64(input)/1e8)
				if !RecentTransactions.ContainsTrans(trans.GetSigHash().String()) {
					RecentTransactions.FactoidTransactions = append(RecentTransactions.FactoidTransactions, struct {
						TxID         string
						Hash         string
						TotalInput   string
						Status       string
						TotalInputs  int
						TotalOutputs int
					}{trans.GetSigHash().String(), trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs})
				}
			}
		} else if entry.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000c" {
			mr := entry.GetKeyMR()

			dbase := StatePointer.GetDB()
			ecblock, err := dbase.FetchECBlock(mr)

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

	if last.GetHeader().GetDBHeight() > RecentTransactions.LastHeightChecked {
		entriesNeeded := 100 - len(RecentTransactions.Entries)
		factoidsNeeded := 100 - len(RecentTransactions.FactoidTransactions)
		// If we do not have 100 of each transaction, we will look into the past to get 100
		if (entriesNeeded + factoidsNeeded) > 0 {
			getPastEntries(last, entriesNeeded, factoidsNeeded)
		} else {
			RecentTransactions.LastHeightChecked = last.GetHeader().GetDBHeight()
		}
	}

	if len(RecentTransactions.Entries) > 100 {
		overflow := len(RecentTransactions.Entries) - 100
		if overflow > 0 {
			RecentTransactions.Entries = RecentTransactions.Entries[overflow:]
		}
	}
	if len(RecentTransactions.FactoidTransactions) > 100 {
		overflow := len(RecentTransactions.FactoidTransactions) - 100
		if overflow > 0 {
			RecentTransactions.FactoidTransactions = RecentTransactions.FactoidTransactions[overflow:]
		}
	}

	// Check if we missed any processing
	for i, e := range RecentTransactions.Entries {
		if e.ChainID == "Processing" {
			entry := getEntry(e.Hash)
			if entry != nil {
				RecentTransactions.Entries[i] = *entry
			}
		}
	}
}

// Control Panel shows the last 100 entry and factoid transactions. This will look into the past if we do not
// currently have 100 of each transaction type. A checkpoint is set each time we check a new height, so we will
// not check a directory block in the past twice.
func getPastEntries(last interfaces.IDirectoryBlock, eNeeded int, fNeeded int) {
	height := last.GetHeader().GetDBHeight()

	next := last.GetHeader().GetPrevKeyMR()
	zero := primitives.NewZeroHash()

	newCheckpoint := height

	for height > RecentTransactions.LastHeightChecked && (eNeeded > 0 || fNeeded > 0) {
		if next.IsSameAs(zero) {
			break
		}
		dbase := StatePointer.GetDB()
		dblk, err := dbase.FetchDBlock(next)

		if err != nil || dblk == nil {
			break
		}
		height = dblk.GetHeader().GetDBHeight()
		ents := dblk.GetDBEntries()
		if len(ents) > 3 && eNeeded > 0 {
			for _, eblock := range ents[3:] {
				dbase := StatePointer.GetDB()
				eblk, err := dbase.FetchEBlock(eblock.GetKeyMR())

				if err != nil || eblk == nil {
					break
				}
				for _, hash := range eblk.GetEntryHashes() {
					if RecentTransactions.ContainsEntry(hash.String()) {
						continue
					}
					e := getEntry(hash.String())
					if e != nil && eNeeded > 0 {
						eNeeded--
						RecentTransactions.Entries = append(RecentTransactions.Entries, *e)
						//RecentTransactions.Entries = append([]EntryHolder{*e}, RecentTransactions.Entries...)
					}
				}
			}
		}
		if fNeeded > 0 {
			fChain := primitives.NewHash(constants.FACTOID_CHAINID)
			for _, entry := range ents {
				if entry.GetChainID().IsSameAs(fChain) {
					dbase := StatePointer.GetDB()
					fblk, err := dbase.FetchFBlock(entry.GetKeyMR())

					if err != nil || fblk == nil {
						break
					}
					transList := fblk.GetTransactions()
					for _, trans := range transList {
						if RecentTransactions.ContainsTrans(trans.GetSigHash().String()) {
							continue
						}
						if trans != nil {
							input, err := trans.TotalInputs()
							if err != nil || input == 0 {
								continue
							}
							totalInputs := len(trans.GetInputs())
							totalOutputs := len(trans.GetECOutputs())
							totalOutputs = totalOutputs + len(trans.GetOutputs())
							inputStr := fmt.Sprintf("%f", float64(input)/1e8)
							fNeeded--
							RecentTransactions.FactoidTransactions = append(RecentTransactions.FactoidTransactions, struct {
								TxID         string
								Hash         string
								TotalInput   string
								Status       string
								TotalInputs  int
								TotalOutputs int
							}{trans.GetSigHash().String(), trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs})
						}
					}
				}
			}
		}
		next = dblk.GetHeader().GetPrevKeyMR()
	}

	DisplayStateMutex.Lock()
	if newCheckpoint < DisplayState.CurrentEBDBHeight && newCheckpoint > RecentTransactions.LastHeightChecked {
		RecentTransactions.LastHeightChecked = newCheckpoint
	}
	DisplayStateMutex.Unlock()
}

// For go routines. Calls function once each duration.
func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func checkControlPanelPassword(response http.ResponseWriter, request *http.Request) bool {
	if false == checkAuthHeader(request) {
		remoteIP := ""
		remoteIP += strings.Split(request.RemoteAddr, ":")[0]
		fmt.Printf("Unauthorized Control Panel client connection attempt from %s\n", remoteIP)
		response.Header().Add("WWW-Authenticate", `Basic realm="factomd Control Panel"`)
		http.Error(response, "401 Unauthorized.", http.StatusUnauthorized)
		return false
	}
	return true
}

func checkAuthHeader(r *http.Request) bool {
	if "" == StatePointer.GetRpcUser() {
		//no username was specified in the config file or command line, meaning factomd control panel is open access
		return true
	}

	authhdr := r.Header["Authorization"]
	if len(authhdr) == 0 {
		return false
	}

	correctAuth := StatePointer.GetRpcAuthHash()

	h := sha256.New()
	h.Write([]byte(authhdr[0]))
	presentedPassHash := h.Sum(nil)

	cmp := subtle.ConstantTimeCompare(presentedPassHash, correctAuth) //compare hashes because ConstantTimeCompare takes a constant time based on the slice size.  hashing gives a constant slice size.
	if cmp != 1 {
		return false
	}
	return true
}
