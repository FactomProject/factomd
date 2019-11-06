//Ͼ/*
// The FactomGenerate templates use Greek Capitol syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package subscribers // this is only here to make gofmt happy and is never in the generated code

//Ͼdefine "subscribe-imports"Ͽ

import (
	"github.com/FactomProject/factomd/pubsub/subscribers"
)

//ϾendϿ

type Ͼ_subscribertypeϿ subscribers.AtomicValue // not used when generating, only used for testing
type Ͼ_valuetypeϿ int                          // not used when generating, only used for testing

// Expects: typename <name> subscribertype <name> valuetype <type>

//Ͼdefine "subscribebyvalue"Ͽ
// Start subscribebyvalue generated go code

// Ͼ_typenameϿ subscriber has the basic necessary function implementations.
type Ͼ_typenameϿ struct {
	Ͼ_subscribertypeϿ
}

func (s *Ͼ_typenameϿ) Value() Ͼ_valuetypeϿ {
	o := s.Ͼ_subscribertypeϿ.Value() // call the generic implementation
	return Ͼ_valuetypeϿ(o)           // cast the return to the specific type
}

// End Subscribebyvalue generated code
//Ͼend Ͽ
