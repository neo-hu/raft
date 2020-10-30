package grpc_transporter

import (
	"context"
	"github.com/neo-hu/raft/logger"
	transporter2 "github.com/neo-hu/raft/transporter"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

type transporter struct {
	mutex    sync.RWMutex
	address  string
	server   *grpc.Server
	peers    map[string]*peer
	delegate transporter2.Delegate
}

func NewGRPCTransporter(address string, opts []grpc.ServerOption) transporter2.Transporter {
	t := &transporter{
		address: address,
		server:  grpc.NewServer(opts...),
		peers:   make(map[string]*peer),
	}
	RegisterTransporterServer(t.server, &handles{transporter: t})
	return t
}

func (t *transporter) Delegate(delegate transporter2.Delegate) {
	t.delegate = delegate
}

func (t *transporter) Start() error {
	l, err := net.Listen("tcp", t.address)
	if err != nil {
		return err
	}
	logger.Infof("transporter grpc listen %s", l.Addr())
	return t.server.Serve(l)
}

func (t *transporter) Stop() error {
	return nil
}

func (t *transporter) PeerCount() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return len(t.peers)
}

func (t *transporter) StartHeartbeat() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, p := range t.peers {
		p.startHeartbeat()
	}
}
func (t *transporter) StopHeartbeat(flush bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	for _, p := range t.peers {
		p.stopHeartbeat(flush)
	}
}

func (t *transporter) AddPeer(name, address string, heartbeatInterval time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.delegate.Name() == name {
		return
	}
	if _, ok := t.peers[name]; ok {
		logger.Debugf("duplicate peer %s", name)
		return
	}
	t.peers[name] = newPeer(name, address, t, heartbeatInterval)

}

func (t *transporter) VoteRequest(ctx context.Context, req *transporter2.RequestVoteRequest) <-chan *transporter2.RequestVoteResponse {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	respChan := make(chan *transporter2.RequestVoteResponse, len(t.peers))
	for _, p := range t.peers {
		go func(p *peer) {
			term, voteGranted, err := p.voteRequest(ctx, req)
			if err == nil {
				respChan <- &transporter2.RequestVoteResponse{
					Term:        term,
					VoteGranted: voteGranted,
				}
			}
		}(p)
	}
	return respChan
}
