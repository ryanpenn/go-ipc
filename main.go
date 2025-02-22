package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/rpc/jsonrpc"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// 文件锁路径
var lockFilePath = "./app.lock"

func main() {
	path, _ := os.Executable()
	_, execName := filepath.Split(path)
	lockFilePath = fmt.Sprintf("%s.lock", strings.TrimSuffix(execName, ".exe"))

	var isNewFile bool
	pf, err := ReadPidFile(lockFilePath)
	if err != nil {
		if errors.Is(err, ErrPidFileNotExists) {
			// 创建 PID 文件
			pf = &PidFile{
				Pid:  os.Getpid(),
				Port: rand.Intn(10000) + 10000, // 随机生成端口号
			}
			isNewFile = true
			if err = WritePidFile(lockFilePath, pf, true); err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("pidfile error:", err)
			os.Exit(1)
		}
	}

	port := pf.Port
	var exists bool
	if !isNewFile {
		// 检查 PID 是否存在
		proc, err := os.FindProcess(pf.Pid)
		if err != nil {
			fmt.Printf("find process error: %s\n", err)
			exists = false
		} else {
			if runtime.GOOS == "windows" {
				exists = true
				pf.Port += rand10()
			} else {
				if err = proc.Signal(syscall.Signal(0)); err != nil {
					exists = true
					pf.Port += rand10()
				}
			}
		}
	} else {
		exists = false
	}

	// 记录当前 PID
	pf.Pid = os.Getpid()
	fmt.Println(pf.Pid, pf.Port, exists, isNewFile)

	app := NewApp(pf.Port)
	// defer app.Stop()
	if exists {
		fmt.Println("syncing data from remote")
		sync(port, app)
		for k, v := range app.Data {
			fmt.Printf("-> app.Data[%s: %v]\n", k, v)
		}
		fmt.Println("synced data from remote")
		write(pf)
		app.Run()
	} else {
		write(pf)
		if err = app.Run(); err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}
}

func write(pf *PidFile) {
	if err := WritePidFile(lockFilePath, pf, true); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func sync(port int, app *App) error {
	client, err := jsonrpc.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Printf("sync error: %s\n", err)
		return err
	}
	defer client.Close()
	return PullData(app, client, 10)
}

func rand10() int {
	n := rand.Intn(10)
	if n == 0 {
		return 1
	}
	if n > 5 {
		return -1 * n
	} else {
		return n
	}
}
