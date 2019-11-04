//Ͼ/*
// The FactomGenerate templates use Canadian Aboriginal syllabary characters using "Ͼ" U+1438, "ᐳ" U+1433 as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//U+03FE	GREEK CAPITAL DOTTED LUNATE SIGMA SYMBOL	Ͼ	view
// U+03FF	GREEK CAPITAL REVERSED DOTTED LUNATE SIGMA SYMBOL	Ͽ
//*/ᐳ

package templates_test // this is only here to make gofmt happy and is never in the generated code

//go:generate go run ./generate.go

//Ͼdefine "accountedqueue_test-imports"ᐳ
import (
	"testing"

	. "templ"

	"github.com/FactomProject/factomd/common"
)

//Ͼendᐳ

// for running the test on the template
type Ͼ_typeᐳ int

var Ͼ_testelementᐳ Ͼ_typeᐳ = 1

//Ͼdefine "accountedqueue_test"ᐳ
// Start accountedqueue_test generated go code

func TestAccountedQueue(t *testing.T) {
	q := new(Ͼ_typenameᐳ).Init(common.NilName, "Test", 10)

	if q.Dequeue() != nil {
		t.Fatal("empty dequeue return non-nil")
	}

	for i := 0; i < 10; i++ {
		q.Enqueue(Ͼ_testelementᐳ)
	}

	// commented out because it requires a modern prometheus package
	//if testutil.ToFloat64(q.TotalMetric()) != float64(10) {
	//	t.Fatal("TotalMetric fail")
	//}

	for i := 9; i >= 0; i-- {
		q.Dequeue()
		// commented out because it requires a modern prometheus package
		//if testutil.ToFloat64(q.Metric()) != float64(i) {
		//	t.Fatal("Metric fail")
		//}
	}

	if q.Dequeue() != nil {
		t.Fatal("empty dequeue return non-nil")
	}
}

//Ͼendᐳ
