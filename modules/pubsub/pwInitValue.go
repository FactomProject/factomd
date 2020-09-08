package pubsub

import (
	"fmt"
	"reflect"
)

var _ IPublisherWrapper = (*PubInitValueWrapper)(nil)

// PubInitValueWrapper will send an initial value when a new subscriber joins
type PubInitValueWrapper struct {
	IPublisher
	PubWrapBase

	update  func(current interface{}, new interface{}) interface{}
	initial interface{}
}

func PubInitMapWrap(initial map[string]interface{}) *PubInitValueWrapper {
	return PubInitValueWrap(initial, MapMaintain)
}

func PubInitValueWrap(initial interface{}, update func(current interface{}, new interface{}) interface{}) *PubInitValueWrapper {
	p := new(PubInitValueWrapper)
	p.initial = initial
	p.update = update

	return p
}

func MapMaintain(current interface{}, new interface{}) interface{} {
	cm, ok := current.(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("Expected initial value to be a map, found %s", reflect.TypeOf(current)))
	}

	// Update the initial
	if m, ok := new.(map[string]interface{}); ok {
		for k, v := range m {
			cm[k] = v
		}
		return cm
	}
	return cm
}

func (m *PubInitValueWrapper) Write(o interface{}) {
	m.initial = m.update(m.initial, o)
	m.IPublisher.Write(o)
}

func (m *PubInitValueWrapper) Wrap(p IPublisher) IPublisherWrapper {
	m.SetBase(p)
	m.IPublisher = p
	return m
}

func (m *PubInitValueWrapper) Subscribe(subscriber IPubSubscriber) bool {
	if m.base.Subscribe(subscriber) {
		subscriber.write(m.initial)
		return true
	}
	return false
}

func (m *PubInitValueWrapper) Publish(path string) IPublisherWrapper {
	if wrap, ok := m.IPublisher.(IPublisherWrapper); ok {
		return wrap.Publish(path)
	}

	globalReg.useLock.Lock()
	defer globalReg.useLock.Unlock()
	globalPublish(path, m)

	return m
}
