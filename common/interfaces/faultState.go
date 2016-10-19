package interfaces

type IFaultState interface {
	IsNil() bool
	HasEnoughSigs(state IState) bool
	String() string
}
