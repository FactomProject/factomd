//Ͼ/*
// The FactomGenerate templates use Greek Capitol syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package threadsafemap // this is only here to make gofmt happy and is never in the generated code
//Ͼdefine "threadsafemap-imports"Ͽ

import (
	"sync"

	"github.com/FactomProject/factomd/common"
)

//ϾendϿ

type Ͼ_indextypeϿ int // not used when generating, only used for testing
type Ͼ_valuetypeϿ int // not used when generating, only used for testing

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
	v, _ := q.GetWithFlag(index)
	return v
}

func (q *Ͼ_typenameϿ) GetWithFlag(index Ͼ_indextypeϿ) (Ͼ_valuetypeϿ, bool) {
	q.Lock()
	defer q.Unlock()
	v, ok := q.internalMap[index]
	return v, ok
}

func (q *Ͼ_typenameϿ) Delete(index Ͼ_indextypeϿ) {
	q.Lock()
	defer q.Unlock()
	delete(q.internalMap, index)
}

func (q *Ͼ_typenameϿ) Len() int {
	q.Lock()
	defer q.Unlock()
	return len(q.internalMap)
}

// End threadsafemap generated go code
//ϾendϿ
