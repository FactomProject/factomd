//Ͼ/*
// The FactomGenerate templates use Greek Capitol syllabary characters "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package accountedqueues // this is only here to make gofmt happy and is never in the generated code

//go:generate go run ../../generate.go

//Ͼdefine "accountedqueue_test-imports"Ͽ

import (
	"testing"

	"github.com/FactomProject/factomd/common"
)

//ϾendϿ

var Ͼ_testelementϿ Ͼ_typeϿ // just use a zero value as the test element

//Ͼdefine "accountedqueue_test"Ͽ
// Start accountedqueue_test generated go code

func TestAccountedQueue_Ͼ_typenameϿ(t *testing.T) {
	q := new(Ͼ_typenameϿ).Init(common.NilName, "TestϾ_typenameϿ", 10)

	if q.Dequeue() != nil {
		t.Fatal("empty dequeue return non-nil")
	}

	for i := 0; i < 10; i++ {
		q.Enqueue(Ͼ_testelementϿ)
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

//ϾendϿ
