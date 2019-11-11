package pubsub_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	. "github.com/FactomProject/factomd/pubsub"
)

func TestRegistry_Subscribe(t *testing.T) {
	ResetGlobalRegistry()
	p := PubFactory.Threaded(5).Publish("test")
	go p.Start()

	s := SubFactory.Channel(5).Subscribe("test")

	data := "random-data"

	go func() {
		p.Write(data)
	}()

	for data2 := range s.Channel() {
		if !reflect.DeepEqual(data, data2) {
			t.Error("Data published is wrong")
		}
		p.Close() // Only 1 read
	}
}

func TestRegistry_BasePublisher(t *testing.T) {
	ResetGlobalRegistry()

	p := PubFactory.Base().Publish("test")
	s := SubFactory.Value().Subscribe("test")

	data := "random-data"

	p.Write(data)
	data2 := s.Read()

	if !reflect.DeepEqual(data, data2) {
		t.Error("Data published is wrong")
	}
}

func TestRegistry_AddPath(t *testing.T) {
	r := NewRegistry()

	r.AddPath(filepath.Join("root", "a", "c"))
	r.AddPath("root/a/d")
	r.AddPath("root/a/f")
	r.AddPath("home/two/f")
	fmt.Println(r.PrintTree())
}
