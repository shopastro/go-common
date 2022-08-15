package globally

import (
	"github.com/gin-gonic/gin"
	"github.com/yousinn/logs"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

type (
	Handler struct {
	}
)

func NewRecoveryHandler() *Handler {
	return &Handler{}
}

func (svc *Handler) RecoveryWithLogger(stack bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			err := recover()
			if err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(ctx.Request, false)
				if brokenPipe {
					logs.Logger.Error(ctx.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)

					_ = ctx.Error(err.(error))
					ctx.Abort()
					return
				}

				if stack {
					logs.Logger.Error(
						"[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.ByteString("stack", debug.Stack()),
					)
				} else {
					logs.Logger.Error(
						"[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}

				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		ctx.Next()
	}
}
