package wsapi_test

import (
	"bytes"
	"encoding/json"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
	"net/http"
	"testing"
)

func TestBaseUrl(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	url := "http://localhost:8088"
	response, err := http.Get(url)
	if err != nil {
		t.Errorf("error: %v", err)
		t.Errorf("response: %v", response)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code: %v", response.StatusCode)
	}

	url = "http://localhost:8088"
	payload, err := json.Marshal("")
	response, err = http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("error: %v", err)
		t.Errorf("response: %v", response)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code: %v", response.StatusCode)
	}
}

// the v2 endpoint ending on a slash with return a 404 not found
func TestTailingSlashes(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	url := "http://localhost:8088/v2/"
	response, err := http.Get(url)
	if err != nil {
		t.Errorf("error: %v", err)
		t.Errorf("response: %v", response)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code: %v", response.StatusCode)
	}
}

// test a method only available for a GET and not a POST
func TestWrongHttpMethod(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	url := "http://localhost:8088/v1/factoid-submit/"
	response, err := http.Get(url)
	if err != nil {
		t.Errorf("error: %v", err)
		t.Errorf("response: %v", response)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code: %v", response.StatusCode)
	}
}
