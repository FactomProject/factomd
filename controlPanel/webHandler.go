package controlpanel

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

type WebHandler interface {
	RegisterRoutes(router *mux.Router)
}

type webHandler struct {
}

func NewWebHandler() WebHandler {
	return &webHandler{}
}

func (handler *webHandler) RegisterRoutes(router *mux.Router) {
	router.NotFoundHandler = http.HandlerFunc(handler.notFound)

	// handle static files
	cssHandler := http.FileServer(http.Dir("resources/css/"))
	jsHandler := http.FileServer(http.Dir("resources/js/"))
	imgHandler := http.FileServer(http.Dir("resources/images/"))
	fontsHandler := http.FileServer(http.Dir("resources/fonts/"))

	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cssHandler))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", jsHandler))
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imgHandler))
	router.PathPrefix("/fonts/").Handler(http.StripPrefix("/fonts/", fontsHandler))

	// register web endpoints
	router.HandleFunc("/", handler.indexHandler)
}

func (handler *webHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle %s '%s' from %s", r.Method, r.URL.Path, r.RemoteAddr)

	t, err := template.ParseFiles("resources/views/index.html")
	if err != nil {
		fmt.Fprintf(w, handler.errorPage(err))
		return
	}

	t.Execute(w, nil)
}

func (handler *webHandler) notFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("page not found %s '%s' from %s", r.Method, r.URL.Path, r.RemoteAddr)
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "custom 404")
}

func (handler *webHandler) errorPage(err error) string {
	return fmt.Sprintf("<h1>%s</h1><div>%v</div>", "error occured", err)
}
