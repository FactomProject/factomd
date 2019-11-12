package debugsettings

import (
	"regexp"
	"sync"

	"github.com/FactomProject/factomd/pubsub"
)

var (
	// All settings per node
	GlobalSettings map[string]*Settings
	setLock        sync.Mutex
)

func init() {
	GlobalSettings = make(map[string]*Settings)
}

// Settings are all the various settings accessible by pub/sub
// All settings are a Mutli Publisher, so they are always accessible.
type Settings struct {
	InputRegEx  pubsub.IPublisher
	OutputRegEx pubsub.IPublisher
	NetStateOff pubsub.IPublisher
}

func GetSettings(nodeName string) *Settings {
	setLock.Lock()
	defer setLock.Unlock()
	return GlobalSettings[nodeName]
}

// NewNode is for node specific settings
func NewNode(nodeName string) *Settings {
	s := new(Settings)

	// Setup publishers
	s.InputRegEx = pubsub.PubFactory.Base().Publish(pubsub.GetPath(nodeName, "settings", "inputregex"), pubsub.PubMultiWrap())
	s.OutputRegEx = pubsub.PubFactory.Base().Publish(pubsub.GetPath(nodeName, "settings", "outputregex"), pubsub.PubMultiWrap())
	s.NetStateOff = pubsub.PubFactory.Base().Publish(pubsub.GetPath(nodeName, "settings", "netstateoff"), pubsub.PubMultiWrap())

	setLock.Lock()
	defer setLock.Unlock()
	GlobalSettings[nodeName] = s
	return s
}

func (s *Settings) UpdateNetStateOff(v bool) {
	s.NetStateOff.Write(v)
}

func (s *Settings) UpdateInputRegex(re string) error {
	v, err := regexp.Compile(re)
	if err != nil {
		return err
	}
	s.InputRegEx.Write(v)
	return nil
}

func (s *Settings) UpdateOuputRegex(re string) error {
	v, err := regexp.Compile(re)
	if err != nil {
		return err
	}
	// *regexp.Regexp{
	s.OutputRegEx.Write(v)
	return nil
}

func (s *Settings) NetStatOffV() *Subscribe_ByValue_Bool_type {
	// The unsafe one does not use locks. We are just changing a boolean value
	return Subscribe_ByValue_Bool(pubsub.SubFactory.UnsafeValue().Subscribe(s.NetStateOff.Path()))
}

func (s *Settings) InputRegexC() *pubsub.SubChannel {
	return pubsub.SubFactory.BEChannel(3).Subscribe(s.InputRegEx.Path())
}
