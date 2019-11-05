//+build ignore

//Ͼ/*
// The FactomGenerate templates use Greek Capitol  syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package templates // this is only here to make gofmt happy and is never in the generated code
//Ͼdefine "threadsafemap-imports"Ͽ

import (
	"sync"

	"github.com/FactomProject/factomd/common"
)

//ϾendϿ

//Ͼdefine "threadsafemap"Ͽ
// Start threadsafemap generated go code

type Ͼ_typenameϿ struct {
	sync.Mutex
	common.Name
	internalMap map[Ͼ_indextypeϿ]Ͼ_valuetypeϿ
}

func (q *Ͼ_typenameϿ) Init(parent common.NamedObject, name string, size int) *Ͼ_typenameϿ {
	q.Name.Init(parent, name)
	q.internalMap = make(map[Ͼ_indextypeϿ]Ͼ_valuetypeϿ, size)
	return q
}

func (q *Ͼ_typenameϿ) Put(index Ͼ_indextypeϿ, value Ͼ_valuetypeϿ) {
	q.Lock()
	q.internalMap[index] = value
	q.Unlock()
}

func (q *Ͼ_typenameϿ) Get(index Ͼ_indextypeϿ) Ͼ_valuetypeϿ {
	q.Lock()
	defer q.Unlock()
	return q.internalMap[index]
}

func (q *Ͼ_typenameᐳ) Len() int {
	q.Lock()
	defer q.Unlock()
	return len(q.internalMap)
}

func (q *Ͼ_typenameᐳ) Cap() int {
	q.Lock()
	defer q.Unlock()
	return cap(q.internalMap)
}

// End threadsafemap generated go code
//ϾendϿ
