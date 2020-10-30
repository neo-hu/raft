package transporter

import (
	"context"
	"fmt"
	"time"
)

type Transporter interface {
	Start() error
	Stop() error
	StopHeartbeat(bool)
	StartHeartbeat()
	Delegate(delegate Delegate)
	AddPeer(name, address string, heartbeatInterval time.Duration)
	PeerCount() int
	VoteRequest(ctx context.Context, req *RequestVoteRequest) <-chan *RequestVoteResponse
}

type RequestVoteRequest struct {
	Term          uint64
	CandidateName string
}

func (r *RequestVoteRequest) String() string {
	return fmt.Sprintf("<RequestVoteRequest CandidateName:%s Term:%d>", r.CandidateName, r.Term)
}

type RequestVoteResponse struct {
	Term        uint64
	VoteGranted bool
}

func (r *RequestVoteResponse) String() string {
	return fmt.Sprintf("<RequestVoteResponse Term:%d VoteGranted:%v>", r.Term, r.VoteGranted)
}

type HeartbeatRequest struct {
	Term uint64
	Name string
}

func (r *HeartbeatRequest) String() string {
	return fmt.Sprintf("<HeartbeatRequest Name:%s Term:%d>", r.Name, r.Term)
}

type HeartbeatResponse struct {
	Term    uint64
	Success bool
	Name    string
}

func (r *HeartbeatResponse) String() string {
	return fmt.Sprintf("<HeartbeatResponse Term:%d Success:%v>", r.Term, r.Success)
}
