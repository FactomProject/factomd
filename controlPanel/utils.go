package controlPanel

import (
	"encoding/json"
)

// Used to send a height as json struct
func HeightToJsonStruct(height uint32) []byte {
	jData, err := json.Marshal(struct{ Height uint32 }{height})
	if err != nil {
		return []byte(`{"Height":0}`)
	}
	return jData
}
