package interfaces

type IElections interface {
	ElectionProcess(IState, IElections)
}