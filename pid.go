package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PidFile struct {
	Pid  int
	Port int
}

var (
	ErrPidFileNotExists  = fmt.Errorf("pidfile not exists")
	ErrPidFileExists     = fmt.Errorf("pidfile exists")
	ErrPidFileInvalid    = fmt.Errorf("pidfile invalid")
	ErrPidFileReadFail   = fmt.Errorf("pidfile read fail")
	ErrPidFileCreateFail = fmt.Errorf("pidfile create fail")
	ErrPidFileWriteFail  = fmt.Errorf("pidfile write fail")
)

func ReadPidFile(filePath string) (*PidFile, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, ErrPidFileNotExists
	}
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, ErrPidFileReadFail
	}
	// 解析文件内容
	var pidFile PidFile
	err = json.Unmarshal(content, &pidFile)
	if err != nil {
		return nil, ErrPidFileInvalid
	}
	return &pidFile, nil
}

func WritePidFile(filePath string, pidFile *PidFile, replace bool) error {
	var pFile *os.File
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，创建文件
		if pFile, err = os.Create(filePath); err != nil {
			return ErrPidFileCreateFail
		}
	}

	if pFile == nil {
		if !replace {
			return ErrPidFileExists
		}

		// 文件存在，打开文件
		var err error
		if pFile, err = os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0666); err != nil {
			return ErrPidFileWriteFail
		}
	}

	// defer 关闭文件
	defer pFile.Close()

	// 写入文件内容
	content, err := json.Marshal(pidFile)
	if err != nil {
		return ErrPidFileWriteFail
	}
	if _, err := pFile.Write(content); err != nil {
		return ErrPidFileWriteFail
	}
	return nil
}
