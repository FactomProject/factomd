package interfaces

type IGeneralMsg interface {
	CreateMsg(messageType byte) IMsg
	UnmarshalMessage(data []byte) (IMsg, error)
	UnmarshalMessageData(data []byte) (newdata []byte, msg IMsg, err error)
	MessageName(Type byte) string
}
