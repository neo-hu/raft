package raft

import (
	"fmt"
	"github.com/neo-hu/raft/logger"
	"github.com/pkg/errors"
	"io"
	"os"
	"sync"
)

type Log struct {
	file       *os.File
	path       string
	entries    []*LogEntry
	mutex      sync.RWMutex
	startIndex uint64
	startTerm  uint64
	server     *server
}

func newLog(server *server) *Log {
	return &Log{
		server:  server,
		entries: make([]*LogEntry, 0),
	}
}

func (l *Log) appendEntry(entry *LogEntry) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.file == nil {
		return errors.New("log is not open")
	}
	if len(l.entries) > 0 {
		lastEntry := l.entries[len(l.entries)-1]
		if entry.Term() < lastEntry.Term() {
			return fmt.Errorf("log: Cannot append entry with earlier term (%x:%x <= %x:%x)", entry.Term(), entry.Index(), lastEntry.Term(), lastEntry.Index())
		} else if entry.Term() == lastEntry.Term() && entry.Index() <= lastEntry.Index() {
			return fmt.Errorf("log: Cannot append entry with earlier index in the same term (%x:%x <= %x:%x)", entry.Term(), entry.Index(), lastEntry.Term(), lastEntry.Index())
		}
	}
	position, _ := l.file.Seek(0, io.SeekCurrent)
	entry.Position = position
	if _, err := entry.Encode(l.file); err != nil {
		return err
	}
	l.entries = append(l.entries, entry)
	return nil
}

func (l *Log) createEntry(term uint64, command Command) (*LogEntry, error) {
	return newLogEntry(l.nextIndex(), term, command)
}

func (l *Log) nextIndex() uint64 {
	return l.currentIndex() + 1
}

func (l *Log) lastInfo() (uint64, uint64) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if len(l.entries) == 0 {
		return l.startIndex, l.startTerm
	}
	entry := l.entries[len(l.entries)-1]
	return entry.Index(), entry.Term()
}

func (l *Log) currentIndex() uint64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.internalCurrentIndex()
}

func (l *Log) internalCurrentIndex() uint64 {
	if len(l.entries) == 0 {
		return l.startIndex
	}
	return l.entries[len(l.entries)-1].Index()
}

func (l *Log) open(path string) error {
	logger.Debugf("log open %s", path)
	var err error
	l.file, err = os.OpenFile(path, os.O_RDWR, 0600)
	l.path = path
	if err != nil {
		if os.IsNotExist(err) {
			// todo 文件不存在，创建
			l.file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
			logger.Debugf("new log %s", path)
			if err != nil {
				return err
			}
		}
		return err
	}
	//var readBytes int64
	// todo 解析日志
	//for {
	//	entry, n, err := newLogEntryFor(l.file, l.server)
	//	if err != nil {
	//		if err != io.EOF {
	//			panic("err")
	//		}
	//		break
	//	}
	//	panic(entry)
	//	readBytes += int64(n)
	//}
	return nil

}
