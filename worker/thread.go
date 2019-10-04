package worker

import (
	"fmt"
)

/*
Defines an interface that we can use to register
coordinated behavior for starting/stopping various parts of factomd


*/
type Thread func(r *Registry, args ...interface{})

type Registry struct {
	Index      int
	Parent     int
	onRun      func()
	onComplete func()
	onExit     func()
}

type RunLevel int

const (
	RUN = iota + 1
	COMPLETE
	EXIT
)

func (r *Registry) Call(level RunLevel) {
	switch level {
	case RUN:
		if r.onRun != nil {
			r.onRun()
		}
	case COMPLETE:
		if r.onComplete != nil {
			r.onComplete()
		}
	case EXIT:
		if r.onExit != nil {
			r.onExit()
		}
	default:
		panic(fmt.Sprintf("unknown runlevel %v", level))
	}
}

func (r *Registry) Run(f func()) *Registry {
	r.onRun = f
	return r
}

func (r *Registry) Complete(f func()) *Registry {
	r.onComplete = f
	return r
}

func (r *Registry) Exit(f func()) *Registry {
	r.onExit = f
	return r
}
