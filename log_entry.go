package raft

import (
	"io"
)

type LogEntry struct {
	Position int64
	index    uint64
	term     uint64
	command  Command
}

func newLogEntry(index uint64, term uint64, command Command) (*LogEntry, error) {
	e := &LogEntry{
		index:   index,
		term:    term,
		command: command,
	}
	return e, nil
}

func newLogEntryFor(r io.Reader, server *server) (*LogEntry, int64, error) {
	panic("22222")
}

func (e *LogEntry) Term() uint64 {
	return e.term
}

func (e *LogEntry) Index() uint64 {
	return e.index
}

func (e *LogEntry) Encode(w io.Writer) (int, error) {
	panic("3333")
}
