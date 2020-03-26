package pubsub

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

type SubWrapMsgFilter struct {
	IPubSubscriber
	SubWrapBase

	whitelist map[byte]struct{}
}

func SubMsgFilterWrap(enabledMsgs ...byte) *SubWrapMsgFilter {
	s := new(SubWrapMsgFilter)
	s.whitelist = make(map[byte]struct{})
	for _, m := range enabledMsgs {
		s.whitelist[m] = struct{}{}
	}

	return s
}

func (s *SubWrapMsgFilter) Wrap(sub IPubSubscriber) IPubSubscriber {
	s.SetBase(sub)
	s.IPubSubscriber = sub
	return s
}

func (s *SubWrapMsgFilter) write(o interface{}) {
	// TODO: Is importing the interfaces package alright here?
	//		We could pass this function in externally.
	msg, ok := o.(interfaces.IMsg)
	if !ok {
		return
	}
	if _, ok := s.whitelist[msg.Type()]; ok {
		s.IPubSubscriber.write(o)
	}
}
