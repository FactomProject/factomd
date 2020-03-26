package controlpanel

import (
	"github.com/FactomProject/factomd/modules/controlPanel/pages"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testWebHandler *webHandler

func init() {
	router := mux.NewRouter()
	indexContent := pages.IndexContent{NodeName: "", BuildNumber: "", Version: ""}
	testWebHandler = &webHandler{IndexPageContent: indexContent}
	testWebHandler.RegisterRoutes(router)
}

func TestControlPanelIndexHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	// e.g. func GetUsersHandler(ctx context.Context, w http.ResponseWriter, r *http.Request)
	handler := http.HandlerFunc(testWebHandler.indexHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
