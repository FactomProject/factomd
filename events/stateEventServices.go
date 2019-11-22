package events

type IStateEventServices interface {
	GetEvents() Events
	IsRunLeader() bool
}
