package globally

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopastro/logs"
	"go.uber.org/zap"
)

type (
	Access struct {
	}
)

func NewGinZap() *Access {
	return &Access{}
}

func (svc *Access) GinZap(timeFormat string, utc bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery
		ctx.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		if len(ctx.Errors) > 0 {
			for _, e := range ctx.Errors.Errors() {
				logs.Logger.Error("[access]", zap.String("message", e))
			}
		} else {
			logs.Logger.Info(path,
				zap.Int("status", ctx.Writer.Status()),
				zap.String("method", ctx.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", ctx.ClientIP()),
				zap.String("user-agent", ctx.Request.UserAgent()),
				zap.String("time", end.Format(timeFormat)),
				zap.Duration("latency", latency),
			)
		}
	}
}
