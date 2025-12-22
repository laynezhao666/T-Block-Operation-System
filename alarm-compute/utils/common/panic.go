package common

import (
	"fmt"
	"runtime"

	"etrpc-go/log"
)

// LogPanic LogPanic
func LogPanic(err interface{}) {
	buf := getStackBuf()
	runtime.Stack(buf, false)
	printPanic(err, buf)
}

// CatchPanic 捕捉panic，打出堆栈信息
func CatchPanic() {
	if err := recover(); err != nil {
		buf := getStackBuf()
		runtime.Stack(buf, false)
		printPanic(err, buf)
	}
}

// CatchPanicCb 捕捉panic，打出堆栈信息
func CatchPanicCb(recoverCb func(interface{})) {
	if err := recover(); err != nil {
		buf := getStackBuf()
		runtime.Stack(buf, false)
		printPanic(err, buf)

		if recoverCb != nil {
			recoverCb(err)
		}
	}
}

func getStackBuf() []byte {
	return make([]byte, 4096)
}

func printPanic(err interface{}, buf []byte) {
	slog := fmt.Sprintf("panic recoverErr: %#v, debug trace: %s", err, buf)
	log.Errorf(slog)
}
