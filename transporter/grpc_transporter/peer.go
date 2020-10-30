package grpc_transporter

import (
	"context"
	"fmt"
	"github.com/neo-hu/raft/logger"
	transporter2 "github.com/neo-hu/raft/transporter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

type peer struct {
	transporter       *transporter
	address           string
	name              string
	connector         func(ctx context.Context) (*grpc.ClientConn, error)
	conn              *grpc.ClientConn
	mutex             sync.RWMutex
	prevLogIndex      uint64
	heartbeatInterval time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
}

func newPeer(name, address string, transporter *transporter, heartbeatInterval time.Duration) *peer {
	gopts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.FailOnNonTempDialError(true),
	}
	connector := func(ctx context.Context) (*grpc.ClientConn, error) {
		conn, err := grpc.DialContext(ctx, address, gopts...)
		return conn, err
	}
	return &peer{
		transporter:       transporter,
		heartbeatInterval: heartbeatInterval,
		address:           address,
		name:              name,
		connector:         connector,
	}
}

func (p *peer) Name() string {
	return fmt.Sprintf("%s[%s]", p.name, p.address)
}
func (p *peer) stopHeartbeat(flush bool) {
	p.cancel()
}
func (p *peer) startHeartbeat() {
	p.ctx, p.cancel = context.WithCancel(context.Background())
	c := make(chan struct{})
	go func() {
		p.heartbeat(c)
	}()
	<-c
}

func (p *peer) heartbeat(c chan struct{}) {
	ctx := p.ctx
	close(c)
	ticker := time.Tick(p.heartbeatInterval)
	logger.Debugf("start peer %s heartbeat %v", p.Name(), p.heartbeatInterval)
	for {
		select {
		case <-ctx.Done():
			logger.Debugf("stop peer %s heartbeat", p.Name())
			return
		case <-ticker:
			ctx, _ := context.WithTimeout(ctx, p.heartbeatInterval)
			p.flush(ctx)
		}
	}
}

func (p *peer) flush(ctx context.Context) {
	resp, err := p.heartbeatRequest(ctx, p.transporter.delegate.Name(), p.transporter.delegate.Term())
	if err != nil {
		//logger.Debugf("heartbeat -> %s error %v", p.Name(), err)
		return
	}
	resp.Name = p.name
	p.transporter.delegate.HeartbeatResponse(resp)
}

func (p *peer) Close() (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.conn != nil {
		err = p.conn.Close()
		p.conn = nil
		return
	}
	return
}

func (p *peer) Client(ctx context.Context) (TransporterClient, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.conn == nil {
		conn, err := p.connector(ctx)
		if err != nil {
			return nil, err
		}
		p.conn = conn
	}
	return NewTransporterClient(p.conn), nil
}

func (p *peer) heartbeatRequest(ctx context.Context, name string, term uint64) (*transporter2.HeartbeatResponse, error) {
	cli, err := p.Client(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Heartbeat(ctx, &HeartbeatRequest{
		Term: term,
		Name: name,
	})
	if err != nil {
		p.unavailableError(err)
		return nil, err
	}
	return &transporter2.HeartbeatResponse{
		Term:    resp.Term,
		Success: resp.Success,
	}, nil
}

func (p *peer) voteRequest(ctx context.Context, req *transporter2.RequestVoteRequest) (uint64, bool, error) {
	cli, err := p.Client(ctx)
	if err != nil {
		return 0, false, err
	}
	resp, err := cli.RequestVote(ctx, &RequestVoteRequest{
		Term:          req.Term,
		CandidateName: req.CandidateName,
	})
	if err != nil {
		p.unavailableError(err)
		return 0, false, err
	}
	return resp.Term, resp.VoteGranted, nil
}

func (p *peer) unavailableError(err error) {
	if status.Code(err) == codes.Unavailable {
		// todo 连接错误，服务不可用
		p.Close()
	}
}

func (p *peer) getPrevLogIndex() uint64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.prevLogIndex
}

func (p *peer) setPrevLogIndex(value uint64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.prevLogIndex = value
}
