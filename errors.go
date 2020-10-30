package raft

import "errors"

var StopError = errors.New("server: Has been stopped")

var NotLeaderError = errors.New("server: Not current leader")
