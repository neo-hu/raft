package transporter

import "context"

type Delegate interface {
	RequestVote(context.Context, *RequestVoteRequest) (*RequestVoteResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	HeartbeatResponse(*HeartbeatResponse)
	Term() uint64
	Name() string
}
