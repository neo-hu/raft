package main

import (
	"fmt"
	"github.com/neo-hu/raft"
	"github.com/neo-hu/raft/transporter/grpc_transporter"
	"time"
)

func main() {
	t1 := grpc_transporter.NewGRPCTransporter(":8181", nil)
	s, err := raft.NewServer("t1", t1)
	if err != nil {
		panic(err)
	}
	t1.AddPeer("t2", "127.0.0.1:8182", raft.DefaultHeartbeatInterval)
	t1.AddPeer("t3", "127.0.0.1:8183", raft.DefaultHeartbeatInterval)
	go func() {
		err = s.Start(nil)
		if err != nil {
			panic(err)
		}
	}()
	t2 := grpc_transporter.NewGRPCTransporter(":8182", nil)
	s2, err := raft.NewServer("t2", t2)
	if err != nil {
		panic(err)
	}
	t2.AddPeer("t1", "127.0.0.1:8181", raft.DefaultHeartbeatInterval)
	t2.AddPeer("t3", "127.0.0.1:8183", raft.DefaultHeartbeatInterval)
	go func() {
		err = s2.Start(nil)
		if err != nil {
			panic(err)
		}
	}()

	//t3 := grpc_transporter.NewGRPCTransporter(":8183", nil)
	//s3, err := raft.NewServer("t3", t3)
	//if err != nil {
	//	panic(err)
	//}
	//t3.AddPeer("t2", "127.0.0.1:8182", raft.DefaultHeartbeatInterval)
	//t3.AddPeer("t1", "127.0.0.1:8181", raft.DefaultHeartbeatInterval)
	//go func() {
	//	time.Sleep(time.Second * 5)
	//	fmt.Println("start t3")
	//	err = s3.Start(nil)
	//	if err != nil {
	//		panic(err)
	//	}
	//}()

	for {
		time.Sleep(time.Second)
		fmt.Println("t1", s.State(), s.Term(), "\t", "t2", s2.State(), s2.Term())
	}
}
