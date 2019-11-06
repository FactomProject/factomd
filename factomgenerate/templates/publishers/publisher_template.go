//Ͼ/*
// The FactomGenerate templates use Greek Capitol syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package publishers

//go:generate go run ../../generate.go

//Ͼdefine "publish-imports"Ͽ

import (
	. "github.com/FactomProject/factomd/common/pubsubtypes"
	. "github.com/FactomProject/factomd/pubsub/publishers"
)

//ϾendϿ

type Ͼ_publishertypeϿ struct{ Base } // not used when generating, only used for testing
type Ͼ_valuetypeϿ DBHT               // not used when generating, only used for testing

// Expects: publishertype <name> valuetype <type>

//Ͼdefine "publish"Ͽ
// Start publisher generated go code

// Publish_Ͼ_publishertypeϿ_Ͼ_valuetypeϿ publisher has the basic necessary function implementations.
type Publish_Ͼ_publishertypeϿ_Ͼ_valuetypeϿ_type struct {
	*Ͼ_publishertypeϿ
}

// Receive the object of type and call the generic so the compiler can check the passed in type
func (p *Publish_Ͼ_publishertypeϿ_Ͼ_valuetypeϿ_type) Write(o Ͼ_valuetypeϿ) {
	p.Ͼ_publishertypeϿ.Write(o)
}

func Publish_Ͼ_publishertypeϿ_Ͼ_valuetypeϿ(p *Ͼ_publishertypeϿ) Publish_Ͼ_publishertypeϿ_Ͼ_valuetypeϿ_type {
	return Publish_Ͼ_publishertypeϿ_Ͼ_valuetypeϿ_type{p}
}

// End publisher generated go code
//ϾendϿ
