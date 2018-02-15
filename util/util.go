package util

import (
	"bytes"
	"encoding/gob"
	//"github.com/FactomProject/electiontesting/imessage"
	//"github.com/FactomProject/electiontesting/messages"
	//"github.com/FactomProject/electiontesting/primitives"
	"reflect"
)

var enc *gob.Encoder
var dec *gob.Decoder

func init() {
	buff := new(bytes.Buffer)
	enc = gob.NewEncoder(buff)
	dec = gob.NewDecoder(buff)
}

// need better reflect based deep copy
func CloneAny(src interface{}) interface{} {
	dst := reflect.New(reflect.TypeOf(src))

	//dst := new(election.Election)
	err := enc.Encode(src)
	if err != nil {
		panic(err)
	}
	err = dec.Decode(&dst)
	if err != nil {
		panic(err)
	}
	return dst
}

//func GetVMForMsg(msg imessage.IMessage, authset primitives.AuthSet, loc primitives.MinuteLocation) int {
//	// If there is no volunteer msg it is not a
//	vol := messages.GetVolunteerMsg(msg)
//	if vol == nil {
//		return -1
//	}
//
//	return authset.VMForIdentity(vol.FaultMsg.Replacing, loc)
//}
