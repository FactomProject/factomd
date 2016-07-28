package controlPanel

import (
	"encoding/json"
	"fmt"
)

// Used to send a height as json struct
func HeightToJsonStruct(height uint32) []byte {
	jData, err := json.Marshal(struct{ Height uint32 }{height})
	if err != nil {
		return []byte(`{"Height":0}`)
	}
	return jData
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		fmt.Println("ERROR: Control Panel has encountered a panic and was halted. Reloading...\n", r)
	}
}
