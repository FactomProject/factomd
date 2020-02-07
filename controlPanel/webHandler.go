package controlpanel

import (
	"fmt"
	"github.com/FactomProject/factomd/controlPanel/pages"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
)

type WebHandler interface {
	RegisterRoutes(router *mux.Router)
}

type webHandler struct {
	IndexPage      pages.Index
	indexTemplate  *template.Template
	searchTemplate *template.Template
}

// NewWebHandler creates a new web handler.
func NewWebHandler(indexPage pages.Index) WebHandler {
	return &webHandler{
		IndexPage: indexPage,
	}
}

// RegisterRoutes initializes the endpoints of the control panel. It registers resources and assigns handlers to requests.
func (handler *webHandler) RegisterRoutes(router *mux.Router) {
	router.NotFoundHandler = http.HandlerFunc(handler.notFound)

	// load templates
	resourceDirectory, err := resourceDirectory()
	if err != nil {
		log.Fatalf("failed to get control panel resource directory: %v", err)
	}

	baseTemplateFile := path.Join(resourceDirectory, "index.html")
	handler.indexTemplate, err = template.ParseFiles(baseTemplateFile, path.Join(resourceDirectory, "views/home.html"))
	if err != nil {
		log.Fatalf("failed to parse control panel index page: %v", err)
	}

	handler.searchTemplate, err = template.ParseFiles(baseTemplateFile, path.Join(resourceDirectory, "views/search.html"))
	if err != nil {
		log.Fatalf("failed to parse control panel search page: %v", err)
	}

	// handle static files
	cssHandler := http.FileServer(http.Dir(filepath.Join(resourceDirectory, "css")))
	jsHandler := http.FileServer(http.Dir(filepath.Join(resourceDirectory, "js")))
	imgHandler := http.FileServer(http.Dir(filepath.Join(resourceDirectory, "images")))
	fontsHandler := http.FileServer(http.Dir(filepath.Join(resourceDirectory, "fonts")))

	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cssHandler))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", jsHandler))
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imgHandler))
	router.PathPrefix("/fonts/").Handler(http.StripPrefix("/fonts/", fontsHandler))

	// register web endpoints
	router.HandleFunc("/", handler.indexHandler)
	router.HandleFunc("/search", handler.searchHandler).Queries("search", "{term}")
}

func (handler *webHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle %s '%s' from %s", r.Method, r.URL.Path, r.RemoteAddr)
	handler.indexTemplate.ExecuteTemplate(w, "site", handler.IndexPage)
}

func (handler *webHandler) searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle %s '%s' from %s", r.Method, r.URL.Path, r.RemoteAddr)

	params := mux.Vars(r)
	term := params["term"]
	log.Printf("%v", params)

	page := pages.Search{
		Title: "Not Found",
		Term:  term,
	}

	err := handler.searchTemplate.ExecuteTemplate(w, "site", page)
	if err != nil {
		log.Printf("failed to render template correctly: %v", err)
	}
}

func (handler *webHandler) notFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("page not found %s '%s' from %s", r.Method, r.URL.Path, r.RemoteAddr)
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "custom 404")
}

func (handler *webHandler) errorPage(err error) string {
	return fmt.Sprintf("<h1>%s</h1><div>%v</div>", "error occured", err)
}

func resourceDirectory() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("no caller information to retrieve control panel resource directory")
		return "", fmt.Errorf("failed to load control panel resource")
	}
	return filepath.Join(path.Dir(filename), "resources"), nil
}
