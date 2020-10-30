package grpc_transporter

import (
	"context"
	transporter2 "github.com/neo-hu/raft/transporter"
)

type handles struct {
	transporter *transporter
}

func (h handles) Heartbeat(ctx context.Context, request *HeartbeatRequest) (*HeartbeatResponse, error) {
	resp, err := h.transporter.delegate.Heartbeat(ctx, &transporter2.HeartbeatRequest{
		Term: request.Term,
		Name: request.Name,
	})
	if err != nil {
		return nil, err
	}
	return &HeartbeatResponse{
		Term:    resp.Term,
		Success: resp.Success,
	}, nil
}

func (h handles) RequestVote(ctx context.Context, request *RequestVoteRequest) (*RequestVoteResponse, error) {
	resp, err := h.transporter.delegate.RequestVote(ctx, &transporter2.RequestVoteRequest{
		Term:          request.Term,
		CandidateName: request.CandidateName,
	})
	if err != nil {
		return nil, err
	}
	return &RequestVoteResponse{
		Term:        resp.Term,
		VoteGranted: resp.VoteGranted,
	}, nil
}
