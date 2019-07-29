package eventmessages

type Event interface {
	Reset()
	String() string
	ProtoMessage()
}
