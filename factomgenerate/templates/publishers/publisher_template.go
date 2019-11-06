//Ͼ/*
// The FactomGenerate templates use Greek Capitol syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package publishers

import (
	"github.com/FactomProject/factomd/pubsub/publishers"
)

//Ͼdefine "publisher-imports"Ͽ

//ϾendϿ

type Ͼ_publishertypeϿ publishers.Base // used when not generating for testing
type Ͼ_valuetypeϿ int                 // used when not generating for testing

// Expects: typename <name> publishertype <name> valuetype <type>

//Ͼdefine "publisher"Ͽ
// Start publisher generated go code

// Ͼ_typenameϿ subscriber has the basic necessary function implementations.
type Ͼ_typenameϿ struct {
	Ͼ_publishertypeϿ
}

// Receive the object of type and call the generic
func (p *Ͼ_typenameϿ) Write(o Ͼ_valuetypeϿ) {
	p.Ͼ_publishertypeϿ.Write(o)
}

// End publisher generated go code
//ϾendϿ
