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

func ServeControlPanel(port int, state *state.State) {
	st = state
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
	case "peers":

	}
}
