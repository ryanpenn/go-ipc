package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

func main() {
	var (
		isProcExists bool
		rpcPort      int
		pf           PidFile
	)
	pc := NewProcessChecker()
	if content, err := pc.ReadPidFile(); err == nil {
		// 读取 PID 文件成功
		if err = json.Unmarshal(content, &pf); err != nil {
			fmt.Printf("pidfile unmarshal error: %s\n", err)
		} else if pc.IsProcessRunning(pf.Pid) {
			// 进程存在
			isProcExists = true
			rpcPort = pf.Port
		}
	}

	pf.Pid = os.Getpid()               // 当前进程 PID
	pf.Port = rand.Intn(10000) + 10000 // 随机生成 RPC 端口号
	content, err := json.Marshal(pf)
	if err != nil {
		fmt.Printf("pidfile marshal error: %s\n", err)
	} else {
		// save pidfile
		if err := pc.WritePidFile(content); err != nil {
			fmt.Printf("pidfile error: %s\n", err)
		}
	}

	if isProcExists {
		// 拉取共享数据
		PullData(rpcPort, 100)
		// 打印共享数据
		fmt.Println("Shared Data:")
		SharedData.Foreach(func(key string, value any) {
			fmt.Printf("pid: %s, port: %v\n", key, value)
		})
	}
	// 当前进程信息
	k, v := strconv.Itoa(pf.Pid), strconv.Itoa(pf.Port)
	SharedData.Set(k, v)
	fmt.Printf("Current Process:\n-> pid: %s, port: %s\n", k, v)

	// rpc server
	svr, err := NewRpcServer(pf.Port)
	if err != nil {
		fmt.Printf("rpc server error: %s\n", err)
		os.Exit(1)
	}
	svr.Register(&SyncService{svr})
	svr.Start()
}
