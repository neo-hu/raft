package main

import (
	"fmt"
	"github.com/neo-hu/raft"
	"github.com/neo-hu/raft/transporter/grpc_transporter"
	"time"
)

func main() {
	t := grpc_transporter.NewGRPCTransporter(":8182", nil)
	s, err := raft.NewServer("t2", t)
	if err != nil {
		panic(err)
	}
	t.AddPeer("t1", "127.0.0.1:8181", raft.DefaultHeartbeatInterval)
	t.AddPeer("t3", "127.0.0.1:8183", raft.DefaultHeartbeatInterval)
	go func() {
		err = s.Start(nil)
		if err != nil {
			panic(err)
		}
	}()
	for {
		time.Sleep(time.Second)
		fmt.Println(s.State(), s.Term(), s.Leader())
	}
}
