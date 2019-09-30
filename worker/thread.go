package worker

/*
Defines an interface that we can use to register
coordinated behavior for starting/stopping various parts of factomd
 */
type Thread func(args ...interface{})