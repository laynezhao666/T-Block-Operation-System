// Package logfile log file
package logfile

import (
	"fmt"
	"agent/entity/consts"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"agent/utils/file"
)

const (
	DefaultLogFileDir  string = "./log/"
	DefaultLogFileName string = "server.log"
	QueryLineMaxNum    int    = 100
)

// QueryServerLogFromFile 查询服务日志
func QueryServerLogFromFile(fileName string, lineNum int) (string, error) {
	if fileName == "" {
		date := time.Now().Format("20060102")
		fileName = DefaultLogFileDir + DefaultLogFileName + date
	}
	exist, _ := file.TestExist(fileName)
	if !exist {
		fileName = DefaultLogFileDir + DefaultLogFileName
		if exist, _ = file.TestExist(fileName); !exist {
			return "", fmt.Errorf("log file not exist")
		}
	}
	return QueryLogFromFile(fileName, lineNum)
}

// QueryLogFromFile 查询日志
func QueryLogFromFile(fileName string, lineNum int) (string, error) {
	exist, _ := file.TestExist(fileName)
	if !exist {
		return "", fmt.Errorf("log file not exist")
	}
	if lineNum <= 0 || lineNum > QueryLineMaxNum {
		lineNum = QueryLineMaxNum
	}
	cmd := exec.Command("tail", "-n", strconv.Itoa(lineNum), fileName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetPacketLogPath 获取日志路径
func GetPacketLogPath(oriCom string) string {
	return consts.ChannelLogDir + "/" + strings.Replace(oriCom, "/", "0x2f", -1)
}
