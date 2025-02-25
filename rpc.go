package main

import (
	"fmt"
	"maps"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Request struct {
	Done bool
	Size int
}

type Reply struct {
	Done bool
	Data map[string]any
}

type SyncService struct {
	svr *RpcServer
}

func (s *SyncService) SyncData(r *Request, reply *Reply) error {
	if r.Done {
		s.svr.Stop() // stop rpc server
		return nil
	}

	size := min(r.Size, SharedData.Len())
	reply.Data = make(map[string]any, size)
	SharedData.Transfer(func(key string, value any) {
		reply.Data[key] = value
		size--
		if size <= 0 {
			return
		}
	})

	reply.Done = len(SharedData.Data) == 0
	return nil
}

type RpcServer struct {
	listener net.Listener
}

func NewRpcServer(port int) (*RpcServer, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &RpcServer{
		listener: l,
	}, nil
}

func (s *RpcServer) Start() {
	fmt.Printf("RPC server start at port: %d\n", s.listener.Addr().(*net.TCPAddr).Port)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			//fmt.Printf("accept error: %s\n", err)
			break
		}

		go func() {
			defer conn.Close()
			jsonrpc.ServeConn(conn)
		}()
	}
}

func (s *RpcServer) Stop() {
	s.listener.Close()
	fmt.Printf("RPC server stopped\n")
}

func (s *RpcServer) Register(rcvr any) error {
	return rpc.Register(rcvr)
}

func PullData(port int, batchSize int) error {
	cli, err := jsonrpc.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Printf("sync error: %s\n", err)
		return err
	}
	defer cli.Close()

	const method = "SyncService.SyncData"
	var reply Reply
	r := &Request{Done: false, Size: batchSize}
	for !reply.Done {
		// call rpc method: syncdata
		if err := cli.Call(method, r, &reply); err != nil {
			fmt.Printf("syncdata error: %s\n", err)
			cli.Call(method, &Request{Size: 0, Done: true}, &reply) // call rpc method: done
			return err
		}
		maps.Copy(SharedData.Data, reply.Data)
	}

	// call rpc method: done
	if err := cli.Call(method, &Request{Size: 0, Done: true}, &reply); err != nil {
		return err
	}
	return nil
}
