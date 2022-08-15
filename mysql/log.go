package mysql

import (
	"context"
	"github.com/yousinn/logs"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"time"
)

type (
	Logger struct {
	}
)

func NewLogger() *Logger {
	return &Logger{}
}

func (log *Logger) LogMode(logger.LogLevel) logger.Interface {
	return &Logger{}
}

func (log *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	logs.Logger.Info(msg, zap.Any("data", data))
}

func (log *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logs.Logger.For(ctx).Warn(msg, zap.Any("data", data))
}

func (log *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	logs.Logger.For(ctx).Error(msg, zap.Any("data", data))
}

func (log *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {

}
