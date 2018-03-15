package imessage

type IMessage interface {
	String() string
	ReadString(s string)
}

type Taggable interface {
	Tag() [32]byte
	TagMessage(tag [32]byte)
}
