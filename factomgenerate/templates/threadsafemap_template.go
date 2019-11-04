//+build ignore

//Ͼ/*
// The FactomGenerate templates use Canadian Aboriginal syllabary characters using "Ͼ" U+1438, "ᐳ" U+1433 as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/ᐳ

package templates // this is only here to make gofmt happy and is never in the generated code
//Ͼdefine "threadsafemap-imports"ᐳ

import (
	"sync"

	"github.com/FactomProject/factomd/common"
)

//Ͼendᐳ

//Ͼdefine "threadsafemap"ᐳ
// Start threadsafemap generated go code

type Ͼ_typenameᐳ struct {
	sync.Mutex
	common.Name
	internalMap map[Ͼ_indextypeᐳ]Ͼ_valuetypeᐳ
}

func (q *Ͼ_typenameᐳ) Init(parent common.NamedObject, name string, size int) *Ͼ_typenameᐳ {
	q.Name.Init(parent, name)
	q.internalMap = make(map[Ͼ_indextypeᐳ]Ͼ_valuetypeᐳ, size)
	return q
}

func (q *Ͼ_typenameᐳ) Put(index Ͼ_indextypeᐳ, value Ͼ_valuetypeᐳ) {
	q.Lock()
	q.internalMap[index] = value
	q.Unlock()
}

func (q *Ͼ_typenameᐳ) Get(index Ͼ_indextypeᐳ) Ͼ_valuetypeᐳ {
	q.Lock()
	defer q.Unlock()
	return q.internalMap[index]
}

// End threadsafemap generated go code
//Ͼendᐳ
