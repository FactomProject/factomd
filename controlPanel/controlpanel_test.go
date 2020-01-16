package controlpanel

import (
	"fmt"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/testHelper"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// test for live testing
func testControlPanel(t *testing.T) {
	// register the fnode so it can be retrieved
	s := testHelper.CreateEmptyTestState()
	fnode.New(s)

	// register the publisher to start the control panel
	p := pubsub.PubFactory.Threaded(5).Publish("test")
	go p.Start()

	go func() {
		i := 1
		for {
			p.Write(fmt.Sprintf("data: %d", i))
			time.Sleep(2 * time.Second)
			i++
		}
	}()

	New(s.FactomNodeName)

	select {}
}

func TestControlPanelIndexHandler(t *testing.T) {
	webHandler := webHandler{}
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	// e.g. func GetUsersHandler(ctx context.Context, w http.ResponseWriter, r *http.Request)
	handler := http.HandlerFunc(webHandler.indexHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
