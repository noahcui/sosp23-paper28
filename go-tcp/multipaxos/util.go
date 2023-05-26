package multipaxos

import (
	pb "github.com/sosp23/replicated-store/go/multipaxos/network"
)

const (
	IdBits               = 0xff
	RoundIncrement       = IdBits + 1
	MaxNumPeers    int64 = 0xf
)

type Peer struct {
	Id   int64
	Stub *pb.TcpLink
	Addr string
}

func MakePeer(addr string, channels *pb.ChannelMap) *pb.TcpLink {
	return pb.NewTcpLink(addr, channels)
}

type ResultType int

const (
	Ok ResultType = iota
	Retry
	SomeElseLeader
)

type Result struct {
	Type   ResultType
	Leader int64
}

func ExtractLeaderId(ballot int64) int64 {
	return ballot & IdBits
}

func IsLeader(ballot int64, id int64) bool {
	return ExtractLeaderId(ballot) == id
}

func IsSomeoneElseLeader(ballot int64, id int64) bool {
	return !IsLeader(ballot, id) && ExtractLeaderId(ballot) < MaxNumPeers
}
