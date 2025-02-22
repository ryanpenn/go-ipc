package main

import (
	"fmt"
	"maps"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"time"
)

type SyncService struct {
	app    *App
	stopCh chan struct{}
}

type Shared struct {
	Size int
	Done bool
}

type Reply struct {
	Done bool
	Data map[string]any
}

func NewSyncService(app *App) *SyncService {
	return &SyncService{app: app, stopCh: make(chan struct{})}
}

func (s *SyncService) start() error {
	// 注册服务
	rpc.Register(s)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.app.Port))
	if err != nil {
		return err
	}

	fmt.Printf("服务启动成功, 监听端口: %d\n", s.app.Port)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept error: %s\n", err)
		}

		go func() {
			defer conn.Close() // 关闭连接
			jsonrpc.ServeConn(conn)
		}()
	}()

	for {
		select {
		case <-s.stopCh:
			fmt.Println("服务停止")
			listener.Close()
			return nil
		case t := <-time.After(time.Second):
			fmt.Println(t.Format(time.DateTime), ":", "服务运行中")
			s.app.Data[strconv.Itoa(os.Getpid())] = len(s.app.Data)
		}
	}
}

func (s *SyncService) Stop() error {
	close(s.stopCh)
	return nil
}

// 发送数据
func (s *SyncService) SyncData(args *Shared, reply *Reply) error {
	fmt.Println("收到同步请求")

	if args.Done {
		reply.Done = true
		s.Stop() // 停止服务
		return nil
	}

	if len(s.app.Data) > 0 {
		count := 0
		reply.Data = make(map[string]any)
		deleteKeys := make([]string, 0)
		for k, v := range s.app.Data {
			deleteKeys = append(deleteKeys, k)
			reply.Data[k] = v
			count++
			if count >= args.Size {
				break
			}
		}
		for _, k := range deleteKeys {
			delete(s.app.Data, k)
		}
	}
	reply.Done = len(s.app.Data) == 0
	return nil
}

// 拉取数据
func PullData(app *App, cli *rpc.Client, batchSize int) error {
	fmt.Println("开始拉取数据")

	const method = "SyncService.SyncData"
	var reply Reply
	args := &Shared{Size: batchSize, Done: false}
	for !reply.Done {
		// 同步数据
		if err := cli.Call(method, args, &reply); err != nil {
			fmt.Printf("同步数据失败: %s\n", err)
			cli.Call(method, &Shared{Size: 0, Done: true}, &reply) // 通知服务停止
			return err
		}
		maps.Copy(app.Data, reply.Data)
	}
	fmt.Println("数据同步完成")

	// 通知服务停止
	if err := cli.Call(method, &Shared{Size: 0, Done: true}, &reply); err != nil {
		return err
	}

	return nil
}
