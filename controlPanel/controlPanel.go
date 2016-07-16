package controlPanel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/FactomProject/factomd/state"
)

var TEMPLATE_PATH string = "./controlPanel/Web/templates/"
var templates = template.Must(template.ParseGlob(TEMPLATE_PATH + "general/*.html")) //Cache general templates

var INDEX_HTML []byte
var mux *http.ServeMux
var st *state.State
var index int = 0
var fnodes []*state.State

func ServeControlPanel(port int, states []*state.State) {
	st = states[index]
	fnodes = states
	portStr := ":" + strconv.Itoa(port)
	fmt.Println("Starting Control Panel on http://localhost" + portStr + "/")
	// Mux for static files
	mux = http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./ControlPanel/Web")))

	INDEX_HTML, _ = ioutil.ReadFile("./ControlPanel/Web/index.html")

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

func factomdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	item := r.FormValue("item") // Item wanted
	switch item {
	case "myHeight":
		data := fmt.Sprintf("%d", st.GetHighestKnownBlock())
		w.Write([]byte(data)) // Return current node height
	case "leaderHeight":
		data := fmt.Sprintf("%d", st.GetLeaderHeight())
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
		data := getRecentTransactions()
		w.Write(data)
	}
}

func getPeers() []byte {
	return []byte("")
}

func getRecentTransactions() []byte {
	last := st.GetDirectoryBlock()
	if last == nil {
		return []byte(`{"list":"none"}`)
	}
	holder := new(LastDirectoryBlockTransactions)
	if holder == nil {
		return []byte(`{"list":"none"}`)
	}

	holder.DirectoryBlock = struct {
		KeyMR     string
		BodyKeyMR string
		FullHash  string
		DBHeight  string

		PrevFullHash string
		PrevKeyMR    string
	}{last.GetKeyMR().String(), last.BodyKeyMR().String(), last.GetFullHash().String(), fmt.Sprintf("%d", last.GetDatabaseHeight()), last.GetHeader().GetPrevFullHash().String(), last.GetHeader().GetPrevKeyMR().String()}

	entries := last.GetDBEntries()
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
				holder.FactoidTransactions = append(holder.FactoidTransactions, struct {
					TxID         string
					TotalInput   string
					Status       string
					TotalInputs  int
					TotalOutputs int
				}{trans.GetHash().String(), inputStr, "Confirmed", totalInputs, totalOutputs})
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
						holder.Entries = append(holder.Entries, *e)
					}
				}
			}
		}
	}

	ret, err := json.Marshal(holder)
	if err != nil {
		return []byte(`{"list":"none"}`)
	}
	return ret
}
