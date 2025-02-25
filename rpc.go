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
		s.svr.Stop() // 停止服务
		return nil
	}

	size := r.Size
	if size > SharedData.Len() {
		size = SharedData.Len()
	}
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
	fmt.Printf("RPC服务启动成功, 监听端口: %d\n", s.listener.Addr().(*net.TCPAddr).Port)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			//fmt.Printf("accept error: %s\n", err)
			break
		}

		go func() {
			defer conn.Close() // 关闭连接
			jsonrpc.ServeConn(conn)
		}()
	}
}

func (s *RpcServer) Stop() {
	s.listener.Close()
	fmt.Printf("RPC服务停止\n")
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
		// 同步数据
		if err := cli.Call(method, r, &reply); err != nil {
			fmt.Printf("同步数据失败: %s\n", err)
			cli.Call(method, &Request{Size: 0, Done: true}, &reply) // 通知服务停止
			return err
		}
		maps.Copy(SharedData.Data, reply.Data)
	}

	// 通知服务停止
	if err := cli.Call(method, &Request{Size: 0, Done: true}, &reply); err != nil {
		return err
	}
	return nil
}
