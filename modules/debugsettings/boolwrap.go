package debugsettings

import (
	"github.com/FactomProject/factomd/pubsub"
)

// Value subscriber has the basic necessary function implementations. All this does is add a wrapper with typing.
type Subscribe_ByValue_Bool_type struct {
	*pubsub.UnsafeSubValue
}

// type the Read function
func (s *Subscribe_ByValue_Bool_type) Read() bool {
	s.UnsafeSubValue.Read()
	return s.UnsafeSubValue.Read().(bool) // cast the return to the specific type
}

// Create a typed instance form a generic instance
func Subscribe_ByValue_Bool(p *pubsub.UnsafeSubValue) *Subscribe_ByValue_Bool_type {
	return &Subscribe_ByValue_Bool_type{p}
}
