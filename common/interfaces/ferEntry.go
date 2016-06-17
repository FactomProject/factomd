package interfaces

type IFEREntry interface {
	GetVersion() string
	SetVersion(passedVersion string) IFEREntry
	GetExpirationHeight() uint32
	SetExpirationHeight(passedExpirationHeight uint32) IFEREntry
	GetResidentHeight() uint32
	SetResidentHeight(passedGetResidentHeight uint32) IFEREntry
	GetTargetActivationHeight() uint32
	SetTargetActivationHeight(passedTargetActivationHeight uint32) IFEREntry
	GetPriority() uint32
	SetPriority(passedPriority uint32) IFEREntry
	GetTargetPrice() uint64
	SetTargetPrice(passedTargetPrice uint64) IFEREntry
}
