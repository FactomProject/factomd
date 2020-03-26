package pubsub

import (
	"reflect"
	"sync"
)

// SubConfig handles a set of atomic values of the last write.
// It is ideal for maintaining a set of configs
type SubConfig struct {
	SubBase
	values map[string]interface{}
	update func(map[string]interface{})
	sync.RWMutex
}

// NewSubConfig instantiates a new value set subscriber. The update
// function can be provided to be called if a change was made.
// So if a new set of fields contains a value that is new, the update
// is called.
func NewSubConfig(update func(map[string]interface{})) *SubConfig {
	s := new(SubConfig)
	s.values = make(map[string]interface{})
	s.update = update

	return s
}

// Pub Side

func (s *SubConfig) write(o interface{}) {
	s.Lock()
	set, ok := o.(map[string]interface{})
	if !ok {
		return
	}

	updated := false
	for k, v := range set {
		orig, ok := s.values[k]
		// If the user wants updates, and we didn't have a field updated
		// yet, check if the value is different than what we have
		if s.update != nil && (updated || !ok || !reflect.DeepEqual(orig, v)) {
			updated = true
		}
		s.values[k] = v
	}
	s.Unlock()
	if s.update != nil && updated {
		s.update(set)
	}
}

// Sub Side

func (s *SubConfig) Read() map[string]interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.values
}

func (s *SubConfig) Subscribe(path string, wrappers ...ISubscriberWrapper) *SubConfig {
	globalSubscribeWith(path, s, wrappers...)
	return s
}

// Accessors
func (s *SubConfig) Field(field string) interface{} {
	return s.values[field]
}

func (s *SubConfig) Int(field string) (int, bool) {
	v, ok := s.values[field].(int)
	return v, ok
}

func (s *SubConfig) String(field string) (string, bool) {
	v, ok := s.values[field].(string)
	return v, ok
}
