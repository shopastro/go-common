package globally

import (
	"github.com/yousinn/logs"
	"go.uber.org/zap"
	"runtime"
	"runtime/debug"
	"time"
)

var Signal = make(chan PanicInformation)

type PanicInformation struct {
	RecoveredPanic interface{}
	Stack          string
}

func Bubble(err interface{}, traceback ...string) {
	if len(traceback) == 0 {
		stack := make([]byte, 1024*8)
		stack = stack[:runtime.Stack(stack, false)]
		traceback = []string{string(stack)}
	}

	Signal <- PanicInformation{
		RecoveredPanic: err,
		Stack:          traceback[0],
	}
}

func Recovers() {
	if err := recover(); err != nil {
		Bubble(err)
	}
}

func AwaitPanics() {
	var pi PanicInformation
	for {
		pi = <-Signal

		logs.Logger.Error("[Recovery from panic]",
			zap.Time("time", time.Now()),
			zap.Any("error", pi.RecoveredPanic),
			zap.String("stack", pi.Stack),
			zap.String("stack", string(debug.Stack())),
		)
	}
}
