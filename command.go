package raft

type Command interface {
	CommandName() string
}

type ApplyCommand interface {
	Apply(Server)
}
