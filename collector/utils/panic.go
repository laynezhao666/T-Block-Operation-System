package utils

import (
	"runtime"

	"etrpc-go/log"
)

// HandlePanic 处理panic并记录堆栈信息
func HandlePanic(panicFrom string) {
	if r := recover(); r != nil {
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, true)
		log.Errorf("%v panic:%v,stack:%s",
			panicFrom, r, string(stack[:length]))
	}
}
