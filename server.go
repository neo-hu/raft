package raft

import (
	"context"
	"github.com/neo-hu/raft/logger"
	transporter2 "github.com/neo-hu/raft/transporter"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultHeartbeatInterval = 300 * time.Millisecond
	//DefaultElectionTimeout   = 150 * time.Millisecond
	DefaultElectionTimeout = 1 * time.Second
)

type Server interface {
	Start(stopCh <-chan struct{}) error
	State() State
	Term() uint64
	Leader() string
	OnEvent(EventHandles)
}

type server struct {
	*eventListener
	mutex       sync.RWMutex
	name        string
	path        string
	transporter transporter2.Transporter
	state       State
	runningFlag int32
	stoppedCh   chan struct{}
	currentTerm uint64
	votedFor    string
	leader      string

	event             chan *deferred
	electionTimeout   time.Duration
	heartbeatInterval time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(name string, t transporter2.Transporter) (Server, error) {
	if name == "" {
		return nil, errors.New("server: name required")
	}
	if t == nil {
		return nil, errors.New("server: transporter required")
	}
	s := &server{
		eventListener:     newEventListener(),
		event:             make(chan *deferred, 256),
		stoppedCh:         make(chan struct{}),
		name:              name,
		state:             Stopped,
		electionTimeout:   DefaultElectionTimeout,
		heartbeatInterval: DefaultHeartbeatInterval,
		transporter:       t,
	}
	t.Delegate(s)
	s.transporter = t
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s, nil
}

func (s *server) ElectionTimeout() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.electionTimeout
}

func (s *server) Running() bool {
	return atomic.LoadInt32(&s.runningFlag) == 1
}
func (s *server) init() error {
	//err := os.Mkdir(path.Join(s.path, "snapshot"), 0700)
	//if err != nil && !os.IsExist(err) {
	//	return fmt.Errorf("raft: Initialization error: %s", err)
	//}

	//err = s.log.open(path.Join(s.path, "log"))
	//if err != nil {
	//	return errors.Wrap(err, "open log")
	//}
	return nil
}

func (s *server) setStateUnlock(state State) {
	logger.Debugf("change state %s => %s", s.state, state)
	prevLeader := s.leader
	s.state = state
	if state == Leader {
		s.leader = s.name
	}
	if prevLeader != s.leader {
		s.dispatchEvent(&leaderChangeEvent{prevLeader: prevLeader, leader: s.leader})
	}
}
func (s *server) setState(state State) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.setStateUnlock(state)

}

func (s *server) State() State {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state
}

func (s *server) followerLoop() {
	timeoutChan := time.After(afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2))
	for s.State() == Follower {
		update := false
		//var err error
		select {
		case <-s.stoppedCh:
			panic("22")
			return
		case e := <-s.event:
			switch req := e.req.(type) {
			case *transporter2.HeartbeatRequest:
				var resp *transporter2.HeartbeatResponse
				resp, update = s.processHeartbeatRequest(req)
				e.done(resp, nil)
			case *transporter2.RequestVoteRequest:
				var resp *transporter2.RequestVoteResponse
				resp, update = s.processRequestVoteRequest(req)
				e.done(resp, nil)
			default:
				panic(req)
				e.done(nil, NotLeaderError)
			}
		case <-timeoutChan:
			s.setState(Candidate)
			//if s.log.currentIndex() > 0 {
			//	s.setState(Candidate)
			//} else {
			//	update = true
			//}
		}
		if update {
			timeoutChan = time.After(afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2))
		}
	}
}

func (s *server) leaderLoop() {
	//logIndex, _ := s.log.lastInfo()
	//logger.Debugf("leaderLoop logIndex %d", logIndex)
	s.transporter.StartHeartbeat()
	atomic.AddUint64(&s.currentTerm, 1)
	for s.State() == Leader {
		select {
		case <-s.stoppedCh:
			s.transporter.StopHeartbeat(false)
			s.setState(Stopped)
			return
		case e := <-s.event:
			switch req := e.req.(type) {
			case *transporter2.HeartbeatRequest:
				resp, _ := s.processHeartbeatRequest(req)
				e.done(resp, nil)
			case *transporter2.HeartbeatResponse:
				s.processHeartbeatResponse(req)
				e.done(nil, nil)
			case *transporter2.RequestVoteRequest:
				resp, _ := s.processRequestVoteRequest(req)
				e.done(resp, nil)
			default:
				panic(req)
			}
		}
	}
}
func (s *server) candidateLoop() {
	prevLeader := s.leader

	s.leader = ""
	if prevLeader != s.leader {
		s.dispatchEvent(&leaderChangeEvent{prevLeader: prevLeader, leader: s.leader})
	}

	var timeoutChan <-chan time.Time
	doVote := true
	//lastLogIndex, lastLogTerm := s.log.lastInfo()
	votesGranted := 0
	var respChan <-chan *transporter2.RequestVoteResponse
	for s.State() == Candidate {
		if doVote {
			s.currentTerm++
			//  todo 开始投票
			t := afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2)
			timeoutChan = time.After(t)
			ctx, _ := context.WithTimeout(context.Background(), t)
			respChan = s.transporter.VoteRequest(ctx, &transporter2.RequestVoteRequest{
				Term:          s.currentTerm,
				CandidateName: s.name,
			})
			votesGranted = 1
			doVote = false
		}

		if votesGranted == s.QuorumSize() {
			s.setState(Leader)
			return
		}

		select {
		case <-s.stoppedCh:
			s.setState(Stopped)
			return
		case resp := <-respChan:
			if success := s.processVoteResponse(resp); success {
				logger.Debugf("vote granted: %d %d", votesGranted, s.currentTerm)
				votesGranted++
			}
		case e := <-s.event:
			switch req := e.req.(type) {
			case *transporter2.HeartbeatRequest:
				resp, _ := s.processHeartbeatRequest(req)
				e.done(resp, nil)
			case *transporter2.RequestVoteRequest:
				resp, _ := s.processRequestVoteRequest(req)
				e.done(resp, nil)
			default:
				panic(req)
				e.done(nil, nil)
			}
		case <-timeoutChan:
			doVote = true
		}
	}
}
func (s *server) Name() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.name
}
func (s *server) Leader() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.leader
}
func (s *server) Term() uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.currentTerm
}
func (s *server) QuorumSize() int {
	return (s.transporter.PeerCount()+1)/2 + 1
}

func (s *server) loop() {
	state := s.State()
	for state != Stopped {
		logger.Debugf("loop run => %s", state)
		switch state {
		case Follower:
			s.followerLoop()
		case Candidate:
			s.candidateLoop()
		case Leader:
			s.leaderLoop()
		case Snapshotting:
			panic("222")
		}
		state = s.State()
	}
}

func (s *server) do(ctx context.Context, value interface{}) (interface{}, error) {
	if !s.Running() {
		return nil, StopError
	}
	if ctx == nil {
		ctx = s.ctx
	}
	deferred := newDeferred(ctx, value)
	select {
	case <-ctx.Done():
		return nil, context.Canceled
	case s.event <- deferred:
	case <-s.stoppedCh:
		return nil, StopError
	}
	select {
	case <-ctx.Done():
		return nil, context.Canceled
	case <-s.stoppedCh:
		return nil, StopError
	case <-deferred.ResultCh():
		return deferred.Result()
	}

}

func (s *server) Start(stopCh <-chan struct{}) error {
	if s.Running() {
		return errors.New("server already running")
	}
	err := s.init()
	if err != nil {
		return err
	}
	s.setState(Follower)

	exitCh := make(chan error)
	var once sync.Once
	exitFunc := func(err error) {
		once.Do(func() {
			exitCh <- err
		})
	}
	go func() {
		exitFunc(s.transporter.Start())
	}()
	go func() {
		s.loop()
	}()
	atomic.StoreInt32(&s.runningFlag, 1)

	select {
	case <-stopCh:
	case err = <-exitCh:
		break
	}
	return err
}

func (s *server) updateCurrentTerm(term uint64, leaderName string) error {
	if term < s.currentTerm {
		return errors.New("update is called when term is not larger than currentTerm")
	}
	prevLeader := s.leader
	if s.state == Leader {
		s.transporter.StopHeartbeat(false)
	}
	if s.state != Follower {
		s.setStateUnlock(Follower)
	}
	s.currentTerm = term
	s.leader = leaderName
	s.votedFor = ""
	if prevLeader != s.leader {
		s.dispatchEvent(&leaderChangeEvent{prevLeader: prevLeader, leader: s.leader})
	}
	return nil
}

func (s *server) processVoteResponse(resp *transporter2.RequestVoteResponse) bool {
	if resp.VoteGranted && resp.Term == s.currentTerm {
		return true
	}
	if resp.Term > s.currentTerm {
		logger.Debugf("candidate vote failed")
		s.updateCurrentTerm(resp.Term, "")
	} else {
		logger.Debugf("candidate vote denied")
	}
	return false
}

func (s *server) processHeartbeatRequest(req *transporter2.HeartbeatRequest) (*transporter2.HeartbeatResponse, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.currentTerm > req.Term {
		logger.Debugf("%s heartbeat term:%d error stale term", req.Name, req.Term)
		return &transporter2.HeartbeatResponse{
			Success: false,
			Term:    s.currentTerm,
		}, false
	}
	if req.Term == s.currentTerm {
		if s.state == Candidate {
			s.setStateUnlock(Follower)
		}
		s.leader = req.Name
	} else {
		s.updateCurrentTerm(req.Term, req.Name)
	}
	return &transporter2.HeartbeatResponse{
		Success: true,
		Term:    s.currentTerm,
	}, true
}

func (s *server) processHeartbeatResponse(resp *transporter2.HeartbeatResponse) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if resp.Term > s.currentTerm {
		if s.state == Leader {
			s.transporter.StopHeartbeat(false)
		}
		if s.state != Follower {
			s.setStateUnlock(Follower)
		}
		s.leader = ""
		s.votedFor = ""
		//s.updateCurrentTerm(resp.Term, "")
		return
	}
	if !resp.Success {
		return
	}
}

func (s *server) processRequestVoteRequest(req *transporter2.RequestVoteRequest) (*transporter2.RequestVoteResponse, bool) {
	// todo 接收到投票
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if req.Term < s.currentTerm {
		logger.Debugf("server receive %s Term:%d < currentTerm:%d cause stale term", req.CandidateName, req.Term, s.currentTerm)
		return &transporter2.RequestVoteResponse{
			Term:        s.currentTerm,
			VoteGranted: false,
		}, false
	}
	if req.Term > s.currentTerm {
		err := s.updateCurrentTerm(req.Term, "")
		if err != nil {
			return &transporter2.RequestVoteResponse{
				Term:        s.currentTerm,
				VoteGranted: false,
			}, false
		}
	} else if s.votedFor != "" && s.votedFor != req.CandidateName {
		logger.Debugf("server receive cause duplicate vote: %s already vote for ", req.CandidateName, s.votedFor)
		return &transporter2.RequestVoteResponse{
			Term:        s.currentTerm,
			VoteGranted: false,
		}, false
	}
	//lastIndex, lastTerm := s.log.lastInfo()
	//if lastIndex > req.LastLogIndex || lastTerm > req.LastLogTerm {
	//	logger.Debugf("server receive cause out of date log: %s Index :[%d] [%d] Term :[%d] [%d]",
	//		req.CandidateName, lastIndex, req.LastLogIndex, lastTerm, req.LastLogTerm)
	//	return &transporter2.RequestVoteResponse{
	//		Term:        s.currentTerm,
	//		VoteGranted: false,
	//	}, false
	//}
	logger.Debugf("server receive votes for %s at term %d", req.CandidateName, req.Term)
	s.votedFor = req.CandidateName
	return &transporter2.RequestVoteResponse{
		Term:        s.currentTerm,
		VoteGranted: true,
	}, true
}

func (s *server) Heartbeat(ctx context.Context, req *transporter2.HeartbeatRequest) (*transporter2.HeartbeatResponse, error) {
	ret, err := s.do(ctx, req)
	if err != nil {
		return nil, err
	}
	resp, _ := ret.(*transporter2.HeartbeatResponse)
	return resp, nil
}

func (s *server) HeartbeatResponse(resp *transporter2.HeartbeatResponse) {
	s.do(nil, resp)
}

func (s *server) RequestVote(ctx context.Context, req *transporter2.RequestVoteRequest) (*transporter2.RequestVoteResponse, error) {
	ret, err := s.do(ctx, req)
	if err != nil {
		return nil, err
	}
	resp, _ := ret.(*transporter2.RequestVoteResponse)
	return resp, nil
}
