package testHelper

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/wsapi"
	"net/http"
	"strconv"
	"testing"
)

const testPort = 8080

func GetRespMap(writer http.ResponseWriter) map[string]interface{} {
	j := GetBody(writer)

	if j == "" {
		return nil
	}

	unmarshalled := map[string]interface{}{}
	err := json.Unmarshal([]byte(j), &unmarshalled)
	if err != nil {
		panic(err)
	}
	return unmarshalled
}

func UnmarshalResp(writer http.ResponseWriter, dst interface{}) {
	j := GetBody(writer)

	type rtn struct {
		Response interface{}
		Success  bool
	}
	r := new(rtn)
	r.Response = dst

	err := json.Unmarshal([]byte(j), r)
	if err != nil {
		fmt.Printf("body - %v\n", j)
		panic(err)
	}
}

func UnmarshalRespDirectly(writer http.ResponseWriter, dst interface{}) {
	j := GetBody(writer)

	err := json.Unmarshal([]byte(j), dst)
	if err != nil {
		fmt.Printf("body - %v\n", j)
		panic(err)
	}
}

func GetRespText(writer http.ResponseWriter) string {
	unmarshalled := GetRespMap(writer)
	if unmarshalled["Response"] != nil {
		marshalled, err := json.Marshal(unmarshalled["Response"])
		if err != nil {
			panic(err)
		}
		return string(marshalled)
	} else {
		marshalled, err := json.Marshal(unmarshalled)
		if err != nil {
			panic(err)
		}
		return string(marshalled)
	}
}

func InitTestState() {
	state := CreateAndPopulateTestStateAndStartValidator()
	state.SetPort(testPort)

	if wsapi.Servers == nil {
		wsapi.Servers = make(map[string]*wsapi.Server)
	}

	port := strconv.Itoa(testPort)
	if wsapi.Servers[port] == nil {
		server := wsapi.InitServer(state)
		wsapi.Servers[port] = server
	}
}

func CreateWebContext(t *testing.T, url string) (http.ResponseWriter, *http.Request){
	responseWriter := new(TestResponseWriter)
	request := CreateWebTestGetRequest(t, url)

	InitTestState()

	return responseWriter, request
}

func CreateWebTestGetRequest(t *testing.T, requestUrl string) *http.Request {
	url := fmt.Sprintf("http://test:%d%s", testPort, requestUrl)
	request, err := http.NewRequest("GET", url , nil)
	if err != nil {
		t.Errorf("failed to create test request: %v, %v", request, err)
		t.FailNow()
	}
	return request
}

func CreateWebTestWriter() http.ResponseWriter{
	return new(TestResponseWriter)
}

type TestResponseWriter struct {
	HeaderCode int
	Head       map[string][]string
	Body       string
}

var _ http.ResponseWriter = (*TestResponseWriter)(nil)

func (t *TestResponseWriter) Header() http.Header {
	if t.Head == nil {
		t.Head = map[string][]string{}
	}
	return (http.Header)(t.Head)
}

func (t *TestResponseWriter) WriteHeader(h int) {
	t.HeaderCode = h
}

func (t *TestResponseWriter) Write(b []byte) (int, error) {
	t.Body = t.Body + string(b)
	return len(b), nil
}

func GetBody(writer http.ResponseWriter) string {
	return writer.(*TestResponseWriter).Body
}
