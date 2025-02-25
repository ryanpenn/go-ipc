package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

type PidFile struct {
	Pid  int
	Port int
}

func main() {
	var (
		isProcExists bool
		rpcPort      int
		pf           PidFile
	)
	pc := NewProcessChecker()
	if content, err := pc.ReadPidFile(); err == nil {
		// unmarshal pidfile
		if err = json.Unmarshal(content, &pf); err != nil {
			fmt.Printf("pidfile unmarshal error: %s\n", err)
		} else if pc.IsProcessRunning(pf.Pid) {
			// process is running
			isProcExists = true
			rpcPort = pf.Port
		}
	}

	pf.Pid = os.Getpid()               // current process pid
	pf.Port = rand.Intn(10000) + 10000 // random rpc port
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
		// pull shared data from rpc server
		PullData(rpcPort, 100)
		// print shared data
		fmt.Println("Shared Data:")
		SharedData.Foreach(func(key string, value any) {
			fmt.Printf("pid: %s, port: %v\n", key, value)
		})
	}
	// current process info
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
