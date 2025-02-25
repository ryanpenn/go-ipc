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
	// check lock file exist
	if _, err := os.Stat(sp.lockFilePath); os.IsNotExist(err) {
		return nil, err
	}
	// read lock file
	content, err := os.ReadFile(sp.lockFilePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (sp *ProcessChecker) WritePidFile(content []byte) error {
	var pFile *os.File
	// check lock file exist
	if _, err := os.Stat(sp.lockFilePath); os.IsNotExist(err) {
		// if not exist, create file
		if pFile, err = os.Create(sp.lockFilePath); err != nil {
			return err
		}
	}

	if pFile == nil {
		// if exist, open file for write
		var err error
		if pFile, err = os.OpenFile(sp.lockFilePath, os.O_WRONLY|os.O_TRUNC, 0666); err != nil {
			return err
		}
	}

	// close file
	defer pFile.Close()

	if _, err := pFile.Write(content); err != nil {
		return err
	}
	return nil
}

func (sp *ProcessChecker) IsProcessRunning(pid int) bool {
	// check pid exist
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

	// set current executable name as lock file name
	sp := &ProcessChecker{
		lockFilePath: fpath,
	}
	for _, o := range opt {
		o(sp)
	}
	return sp
}
