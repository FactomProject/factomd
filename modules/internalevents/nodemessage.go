package internalevents

type NodeMessageCode int32

const (
	NodeMessageCode_GENERAL  NodeMessageCode = 0
	NodeMessageCode_STARTED  NodeMessageCode = 1
	NodeMessageCode_SYNCED   NodeMessageCode = 2
	NodeMessageCode_SHUTDOWN NodeMessageCode = 3
)

type Level int32

const (
	Level_INFO    Level = 0
	Level_WARNING Level = 1
	Level_ERROR   Level = 2
)

type NodeMessage struct {
	MessageCode NodeMessageCode
	Level       Level
	MessageText string
}
