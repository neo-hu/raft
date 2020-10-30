package main

import (
	"fmt"
	"github.com/neo-hu/raft"
	"github.com/neo-hu/raft/transporter/grpc_transporter"
	"time"
)

type e struct {
}

func (e *e) OnLeaderChange(leader, prevLeader string) {
	fmt.Printf("leader change %s => %s\n", prevLeader, leader)
}

func main() {
	t := grpc_transporter.NewGRPCTransporter(":8181", nil)
	s, err := raft.NewServer("t1", t)
	if err != nil {
		panic(err)
	}
	t.AddPeer("t2", "127.0.0.1:8182", raft.DefaultHeartbeatInterval)
	t.AddPeer("t3", "127.0.0.1:8183", raft.DefaultHeartbeatInterval)
	s.OnEvent(&e{})
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
