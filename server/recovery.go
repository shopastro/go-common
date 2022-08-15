package server

import (
	"fmt"
	"github.com/yousinn/logs"
	"go.uber.org/zap"
	"runtime"
	"runtime/debug"
	"time"
)

func doLog(p interface{}, caller, stack string) {
	logs.Logger.Error("[Grpc Recovery] [Recovery from panic]",
		zap.Time("time", time.Now()),
		zap.Any("error", p),
		zap.String("caller", caller),
		zap.String("debugStack", stack),
	)
}

func LogRecoverStack(p interface{}) {
	logRecoverStack(p, 4)
}

func logRecoverStack(p interface{}, skip int) {
	var caller string
	if pc, file, line, ok := runtime.Caller(skip); ok {
		caller = fmt.Sprintf("func: %s file: %s:%d", runtime.FuncForPC(pc).Name(), file, line)
	}
	doLog(p, caller, string(debug.Stack()))
}
