package client

import (
	"fmt"

	"github.com/shopastro/logs"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type (
	Log struct {
		ctx context.Context
	}
)

func NewLog(ctx context.Context) *Log {
	return &Log{
		ctx: ctx,
	}
}

func (svc *Log) SetPrefix(prefix string) {
	logs.Logger.Warn("[client logger]", zap.String("prefix", prefix))
}

func (svc *Log) Printf(format string, v ...interface{}) {
	logs.Logger.Warn(fmt.Sprintf("[%s]", format), zap.Any("value", v))
}

func (svc *Log) Println(v ...interface{}) {
	logs.Logger.Warn("[client logger]", zap.Any("value", v))
}
