package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

type ProcessChecker struct {
	lockFilePath string
}

func (sp *ProcessChecker) ReadPidFile() ([]byte, error) {
	// 检查文件是否存在
	if _, err := os.Stat(sp.lockFilePath); os.IsNotExist(err) {
		return nil, err
	}
	// 读取文件内容
	content, err := os.ReadFile(sp.lockFilePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (sp *ProcessChecker) WritePidFile(content []byte) error {
	var pFile *os.File
	// 检查文件是否存在
	if _, err := os.Stat(sp.lockFilePath); os.IsNotExist(err) {
		// 文件不存在，创建文件
		if pFile, err = os.Create(sp.lockFilePath); err != nil {
			return err
		}
	}

	if pFile == nil {
		// 文件存在，打开文件
		var err error
		if pFile, err = os.OpenFile(sp.lockFilePath, os.O_WRONLY|os.O_TRUNC, 0666); err != nil {
			return err
		}
	}

	// defer 关闭文件
	defer pFile.Close()

	if _, err := pFile.Write(content); err != nil {
		return err
	}
	return nil
}

func (sp *ProcessChecker) IsProcessRunning(pid int) bool {
	// 检查 PID 是否存在
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	} else {
		if runtime.GOOS == "windows" {
			return true
		} else {
			if err = proc.Signal(syscall.Signal(0)); err == nil {
				return true
			}
		}
	}
	return false
}

type ProcessOption func(*ProcessChecker)

func WithLockFile(lockFilePath string) ProcessOption {
	return func(sp *ProcessChecker) {
		sp.lockFilePath = lockFilePath
	}
}

func NewProcessChecker(opt ...ProcessOption) *ProcessChecker {
	path, _ := os.Executable()
	_, execName := filepath.Split(path)
	fpath := fmt.Sprintf("%s.lock", strings.TrimSuffix(execName, ".exe"))

	sp := &ProcessChecker{
		lockFilePath: fpath, // 默认锁文件名为可执行文件名加上.lock后缀
	}
	for _, o := range opt {
		o(sp)
	}
	return sp
}
