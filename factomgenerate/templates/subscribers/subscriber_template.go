//Ͼ/*
// The FactomGenerate templates use Greek Capitol syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package subscribers // this is only here to make gofmt happy and is never in the generated code

// THis defines the imports for all the subscription types
//Ͼdefine "subscribe-imports"Ͽ

import (
	. "github.com/FactomProject/factomd/common/pubsubtypes"
	. "github.com/FactomProject/factomd/pubsub"
)

//ϾendϿ
//Ͼdefine "subscribe_byvalue-imports"Ͽ Ͼtemplate "subscribe-imports"Ͽ ϾendϿ  // use the common imports list
//Ͼdefine "subscribe_bychannel-imports"Ͽ Ͼtemplate "subscribe-imports"Ͽ ϾendϿ // use the common imports list

type Ͼ_valuetypeϿ DBHT // not used when generating, only used for testing. insures pubsubtypes in imported

// Expects: valuetype <type>
//Ͼdefine "subscribe_byvalue"Ͽ
// Start subscribeByValue generated go code

// Value subscriber has the basic necessary function implementations. All this does is add a wrapper with typing.
type Subscribe_ByValue_Ͼ_valuetypeϿ_type struct {
	*Value
}

// type the Read function
func (s *Subscribe_ByValue_Ͼ_valuetypeϿ_type) Read() Ͼ_valuetypeϿ {
	return s.Value.Read().(Ͼ_valuetypeϿ) // cast the return to the specific type
}

// Create a typed instance form a generic instance
func Subscribe_ByValue_Ͼ_valuetypeϿ(p *Value) *Subscribe_ByValue_Ͼ_valuetypeϿ_type {
	return &Subscribe_ByValue_Ͼ_valuetypeϿ_type{p}
}

// End subscribe_byvalue generated code
//Ͼend Ͽ

// Expects: valuetype <type>
//Ͼdefine "subscribe_bychannel"Ͽ
// Start subscribeBychannel generated go code

// Channel subscriber has the basic necessary function implementations. All this does is add a wrapper with typing.
type Subscribe_Bychannel_Ͼ_valuetypeϿ_type struct {
	*Channel
}

// type the Read function
func (s *Subscribe_Bychannel_Ͼ_valuetypeϿ_type) Read() Ͼ_valuetypeϿ {
	return s.Channel.Read().(Ͼ_valuetypeϿ) // cast the return to the specific type
}

// type the ReadWithInfo function
func (s *Subscribe_Bychannel_Ͼ_valuetypeϿ_type) ReadWithInfo() (Ͼ_valuetypeϿ, bool) {
	v, ok := <-s.Updates
	return v.(Ͼ_valuetypeϿ), ok
}

// Create a typed instance form a generic instance
func Subscribe_Bychannel_Ͼ_valuetypeϿ(p *Channel) *Subscribe_Bychannel_Ͼ_valuetypeϿ_type {
	return &Subscribe_Bychannel_Ͼ_valuetypeϿ_type{p}
}

// End subscribe_bychannel generated code
//Ͼend Ͽ
