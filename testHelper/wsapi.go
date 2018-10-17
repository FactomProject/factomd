package testHelper

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FactomProject/web"
)

func GetRespMap(context *web.Context) map[string]interface{} {
	j := GetBody(context)

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

func UnmarshalResp(context *web.Context, dst interface{}) {
	j := GetBody(context)

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

func UnmarshalRespDirectly(context *web.Context, dst interface{}) {
	j := GetBody(context)

	err := json.Unmarshal([]byte(j), dst)
	if err != nil {
		fmt.Printf("body - %v\n", j)
		panic(err)
	}
}

func GetRespText(context *web.Context) string {
	unmarshalled := GetRespMap(context)
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

func ClearContextResponseWriter(context *web.Context) {
	context.ResponseWriter = new(TestResponseWriter)
}

func CreateWebContext() *web.Context {
	context := new(web.Context)
	context.Server = new(web.Server)
	context.Server.Env = map[string]interface{}{}
	context.Server.Env["state"] = CreateAndPopulateTestStateAndStartValidator()
	context.ResponseWriter = new(TestResponseWriter)

	return context
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

func GetBody(context *web.Context) string {
	return context.ResponseWriter.(*TestResponseWriter).Body
}
