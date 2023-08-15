package network

type ResponseType int32

const (
	Ok     ResponseType = 0
	Reject ResponseType = 1
)

type CommandType int32

const (
	Get CommandType = 0
	Put CommandType = 1
	Del CommandType = 2
)

type InstanceState int32

const (
	Inprogress InstanceState = 0
	Committed  InstanceState = 1
	Executed   InstanceState = 2
)

type MessageType uint8

const (
	PREPAREREQUEST MessageType = iota + 1
	PREPARERESPONSE
	ACCEPTREQUEST
	ACCEPTRESPONSE
	COMMITREQUEST
	COMMITRESPONSE
	FORWARD
	FORWARDRESPONSE
)

type Command struct {
	ID    string
	Type  CommandType
	Key   string
	Value string
}

type Forward struct {
	ClientId int64
	Command  *Command
}

type ForwardResponse struct {
	Type ResponseType
}

type Instance struct {
	Ballot   int64
	Index    int64
	ClientId int64
	State    InstanceState
	Command  *Command
}

type PrepareRequest struct {
	Ballot int64
	Sender int64
}

type PrepareResponse struct {
	Type   ResponseType
	Ballot int64
	Logs   []*Instance
}

type AcceptRequest struct {
	Instance *Instance
	Sender   int64
}

type AcceptResponse struct {
	Type   ResponseType
	Ballot int64
}

type CommitRequest struct {
	Ballot             int64
	LastExecuted       int64
	GlobalLastExecuted int64
	Sender             int64
}

type CommitResponse struct {
	Type         ResponseType
	Ballot       int64
	LastExecuted int64
}
