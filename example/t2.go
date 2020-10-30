package main

import (
	"fmt"
	"google.golang.org/grpc"
)

func main() {
	gopts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.FailOnNonTempDialError(true),
	}
	conn, err := grpc.Dial("127.0.0.1:8186", gopts...)
	fmt.Println(conn, err)
}
