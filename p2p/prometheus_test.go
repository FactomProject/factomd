package p2p

import (
	"reflect"
	"testing"
)

func TestPrometheus_Setup(t *testing.T) {
	p := new(Prometheus)
	p.Setup()

	vals := reflect.ValueOf(*p)
	for i := 0; i < vals.NumField(); i++ {
		field := vals.Field(i)
		if field.IsNil() {
			t.Errorf("Prometheus.%s %s is nil after Setup()", reflect.TypeOf(*p).Field(i).Name, field.Type().Name())
		}
	}
}
