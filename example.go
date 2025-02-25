package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Args struct {
	A, B int
}

type ExampleService struct{}

func (s *ExampleService) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

// rpc server
func ExampleRpcServer() {
	service := new(ExampleService)
	rpc.Register(service)
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Server started on port 1234")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Failed to accept: %v", err)
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
}

// rpc client
func ExampleRpcClient() {
	client, err := jsonrpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer client.Close()

	args := &Args{7, 8}
	var reply int
	err = client.Call("ExampleService.Multiply", args, &reply)
	if err != nil {
		log.Fatalf("Failed to call: %v", err)
	}
	fmt.Printf("Result: %d\n", reply)
}
