package interfaces

type IMsgFactory interface {
	UnmarshalMessageData(data []byte) (newdata []byte, msg IMsg, err error)
	MessageName(Type byte) string
	SignSignable(s Signable, key Signer) (IFullSignature, error)
	VerifyMessage(s Signable) (bool, error)
}

type Signable interface {
	Sign(Signer) error
	MarshalForSignature() ([]byte, error)
	GetSignature() IFullSignature
	IsValid() bool // Signature already checked
	SetValid()     // Mark as validated so we don't have to repeat.
}

type MultiSignable interface {
	AddSignature(Signer) error
	MarshalForSignature() ([]byte, error)
	GetSignatures() []IFullSignature
	VerifySignatures() ([]IFullSignature, error)
}
