package primitives

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
)

func Log(format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	fmt.Printf(file+":"+strconv.Itoa(line)+" - "+format+"\n", args...)
}

func LogJSONs(format string, args ...interface{}) {
	jsons := []interface{}{}
	for _, v := range args {
		j, _ := encodeJSONString(v)
		jsons = append(jsons, j)
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	fmt.Printf(file+":"+strconv.Itoa(line)+" - "+format+"\n", jsons...)
}

func encodeJSON(data interface{}) ([]byte, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func encodeJSONString(data interface{}) (string, error) {
	encoded, err := encodeJSON(data)
	if err != nil {
		return "", err
	}
	return string(encoded), err
}
