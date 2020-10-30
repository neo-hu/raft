package raft

type State string

const (
	Stopped      State = "stopped"
	Initialized  State = "initialized"
	Follower     State = "follower"
	Candidate    State = "candidate"
	Leader       State = "leader"
	Snapshotting State = "snapshotting"
)
