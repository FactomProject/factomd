package pubregistry_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Emyrk/pubsub/publishers"
	. "github.com/Emyrk/pubsub/pubregistry"
	"github.com/Emyrk/pubsub/subscribers"
)

func TestRegistry_Subscribe(t *testing.T) {
	r := NewRegistry()
	p := publishers.NewSimpleMultiPublish(5)
	err := r.Register("test", p)
	if err != nil {
		t.Error(err)
	}

	s := subscribers.NewChannelBasedSubscriber(5)
	err = r.SubscribeTo("test", s)
	if err != nil {
		t.Error(err)
	}

	data := "random-data"

	go func() {
		p.Write(data)
	}()

	data2 := s.Receive()

	if !reflect.DeepEqual(data, data2) {
		t.Error("Data published is wrong")
	}
}

func TestRegistry_ThreadedPublisher(t *testing.T) {
	r := NewRegistry()
	p := publishers.NewThreadedPublisherPublisher(5)
	go p.Run()
	err := r.Register("test", p)
	if err != nil {
		t.Error(err)
	}

	s := subscribers.NewAtomicValueSubscriber()
	err = r.SubscribeTo("test", s)
	if err != nil {
		t.Error(err)
	}

	data := "random-data"

	go func() {
		p.Write(data)
	}()

	data2 := s.Value()

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
