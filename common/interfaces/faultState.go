package interfaces

type IFaultState interface {
	IsNil() bool
	HasEnoughSigs(state IState) bool
	String() string

	GetAmINegotiator() bool
	SetAmINegotiator(b bool)
	GetMyVoteTallied() bool
	SetMyVoteTallied(b bool)
	GetNegotiationOngoing() bool
	SetNegotiationOngoing(b bool)
	GetPledgeDone() bool
	SetPledgeDone(b bool)
	GetLastMatch() int64
	SetLastMatch(b int64)
}
