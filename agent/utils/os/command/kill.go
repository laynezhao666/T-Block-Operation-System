package command

import (
	"fmt"
)

const (
	SignalKill = "KILL"
)

// GenerateKillCommand 生成向指定进程发送信号的命令
func GenerateKillCommand(process, signal string) string {
	return fmt.Sprintf(
		"t=$(/bin/ps -o pid,args -e | grep \"%v\" | grep -v grep | grep -v tail | awk '{print $1}') "+
			"&& if [ -n \"${t}\" ]; then kill -s %v ${t}; fi", process, signal,
	)
}

// GenerateKillSigalKillCommand 生成向指定进程发送 SIGKILL 信号的命令
func GenerateKillSigalKillCommand(process string) string {
	return GenerateKillCommand(process, SignalKill)
}
