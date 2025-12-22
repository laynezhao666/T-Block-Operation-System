package utils

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/logic/debug"
)

const (
	filename = "debug-points.txt"
)

var (
	file      *os.File = nil
	fileMutex sync.RWMutex

	newLine = []byte("\n")
)

func init() {
	debug.RegisterDisableHandler(closeFile)
}

func closeFile() {
	if file == nil {
		return
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	log.Infof("close debug file")
	_ = file.Close()
}

func openFile() *os.File {
	fileMutex.RLock()
	if file != nil {
		fileMutex.RUnlock()
		return file
	}
	fileMutex.RUnlock()

	fileMutex.Lock()
	defer fileMutex.Unlock()

	if file != nil {
		return file
	}

	var err error
	if file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
		log.Warnf("open file %v error:%v", filename, err)
		file = nil
		return nil
	}

	return file
}

// DebugRecord 开启 debug 模式后，记录推送的测点数据到本地文件
func DebugRecord(k *KafkaData) {
	go record(k)
}

func record(k *KafkaData) {
	if k == nil {
		return
	}
	if !debug.IsEnable() {
		return
	}

	var f *os.File = nil
	if f = openFile(); f == nil {
		return
	}

	var (
		b   []byte
		err error
	)
	if b, err = json.Marshal(k); err != nil {
		return
	}

	t := time.Now().Format("2006-01-02 15:04:05.000000") + " "

	fileMutex.Lock()
	defer fileMutex.Unlock()

	if _, err = f.WriteString(t); err != nil {
		return
	}
	if _, err = f.Write(b); err != nil {
		return
	}
	_, _ = f.Write(newLine)
}
