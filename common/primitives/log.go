// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
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
		j, _ := EncodeJSONString(v)
		jsons = append(jsons, j)
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	fmt.Printf(file+":"+strconv.Itoa(line)+" - "+format+"\n", jsons...)
}

func DecodeJSON(data []byte, v interface{}) error {
	err := json.Unmarshal(data, &v)
	return err
}

func EncodeJSON(data interface{}) ([]byte, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func EncodeJSONString(data interface{}) (string, error) {
	encoded, err := EncodeJSON(data)
	if err != nil {
		return "", err
	}
	return string(encoded), err
}

func DecodeJSONString(data string, v interface{}) error {
	return DecodeJSON([]byte(data), v)
}

func EncodeJSONToBuffer(data interface{}, b *bytes.Buffer) error {
	encoded, err := EncodeJSON(data)
	if err != nil {
		return err
	}
	_, err = b.Write(encoded)
	return err
}
