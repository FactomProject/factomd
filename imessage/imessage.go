package imessage

type IMessage interface {
	String() string
	ReadString(s string)
}
