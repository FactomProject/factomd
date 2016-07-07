package controlPanel

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/FactomProject/factomd/state"
)

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
	w.Write(INDEX_HTML)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	method := r.FormValue("method")
	switch method {
	case "search":
		fmt.Println(r.FormValue("search"))
		found, respose := searchDB(r.FormValue("search"), st)
		if found {
			w.Write([]byte(respose))
			return
		}
	}
	w.Write([]byte(""))
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
	case "peers":

	}
}
