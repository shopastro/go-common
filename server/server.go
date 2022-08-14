package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/opentracing/opentracing-go"
	"github.com/shopastro/go-common/common"
	"github.com/shopastro/go-common/controller"
	"github.com/shopastro/go-common/globally"
	"github.com/shopastro/go-common/tracer"
	"github.com/shopastro/logs"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

type (
	GinServer struct {
		ServerCfg     *Config
		Engine        *gin.Engine
		tools         *common.Tools
		RouterGroup   *gin.RouterGroup
		I18nBundle    *i18n.Bundle
		LoopCall      func(structs ...interface{})
		RegisterRoute func()
		GrpcServer    *GrpcServer
		Tracer        opentracing.Tracer
		CliCtx        *cli.Context
		Metadata      map[string]string
		registered    bool
		rw            sync.RWMutex
		exit          chan chan error
		Id            string
	}

	Config struct {
		ContextPath string        `json:"contextPath" yaml:"contextPath"`
		Host        string        `json:"host" yaml:"host"`
		Port        int           `json:"port" yaml:"port"`
		GrpcPort    int           `json:"grpcPort" yaml:"grpcPort"`
		Mode        string        `json:"mode" yaml:"mode"`
		Debug       bool          `json:"debug" yaml:"debug"`
		TraceParam  float64       `json:"traceParam" yaml:"traceParam"`
		Namespace   string        `json:"namespace" yaml:"namespace"`
		TTL         time.Duration `json:"ttl" yaml:"ttl"`
		Interval    time.Duration `json:"interval" yaml:"interval"`
		HttpPort    int32         `json:"httpPort" yaml:"httpPort"`
	}
)

const (
	Localizer         = "Localizer"
	Language          = "language"
	HttpPortDefault   = 80
	GrpcPortDefault   = 8080
	defaultHealthPath = "/health"
)

func NewGinServer(cfg *Config, cliCtx *cli.Context) *GinServer {
	if cfg == nil {
		log.Fatal("[The system configuration is empty, please configure the system configuration]")
	}

	gin.SetMode(cfg.Mode)

	return &GinServer{
		tools:      common.NewTools(),
		ServerCfg:  cfg,
		Engine:     gin.New(),
		I18nBundle: i18n.NewBundle(language.Chinese),
		GrpcServer: NewGrpcServer(),
		LoopCall: func(structs ...interface{}) {
			for _, v := range structs {
				classType := reflect.TypeOf(v)
				classValue := reflect.ValueOf(v)

				for i := 0; i < classType.NumMethod(); i++ {
					m := classValue.MethodByName(classType.Method(i).Name)
					if m.IsValid() {
						var params []reflect.Value
						m.Call(params)
					}
				}
			}
		},
		Tracer: tracer.NewTracer(&tracer.TraceConfig{
			Param:       cfg.TraceParam,
			ServiceName: strings.TrimPrefix(cfg.ContextPath, "/"),
		}),
		CliCtx: cliCtx,
		exit:   make(chan chan error),
		Id:     common.NewTools().GetRandomString(8),
	}
}

func (svc *GinServer) Run() {
	svc.Engine.Use(globally.NewRecoveryHandler().RecoveryWithLogger(!svc.ServerCfg.Debug))

	if svc.ServerCfg.Debug {
		svc.Engine.Use(gin.Logger())
	}

	svc.Grpc().LoopCall()
}

func (svc *GinServer) Grpc() *GinServer {
	defer globally.Recovers()

	if svc.GrpcServer.RegisteGrpcServer != nil {
		if svc.ServerCfg.GrpcPort <= 0 {
			svc.ServerCfg.GrpcPort = GrpcPortDefault
		}

		grpcAddr := fmt.Sprintf("%s:%d", svc.ServerCfg.Host, svc.ServerCfg.GrpcPort)
		fmt.Println("grpc server: ", grpcAddr)

		var err error
		svc.GrpcServer.Listener, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			logs.Logger.Fatal("[failed to listen]", zap.Any("error", err))
		}

		go func() {
			logs.Logger.Fatal("[Grpc Server]", zap.Any("error", svc.GrpcServer.RunGrpcServe()))
		}()
	}

	return svc
}

func (svc *GinServer) RunHttpServe() error {
	svc.newRoute().RegisterRoute()

	if svc.ServerCfg.Port <= 0 {
		svc.ServerCfg.Port = HttpPortDefault
	}

	httpAddr := fmt.Sprintf("%s:%d", svc.ServerCfg.Host, svc.ServerCfg.Port)
	return svc.Engine.Run(httpAddr)
}

func (svc *GinServer) newRoute() *GinServer {
	prom := NewPrometheus("gin")
	prom.Use(svc.Engine)

	svc.RouterGroup = svc.Engine.Group(
		svc.ServerCfg.ContextPath,
		tracer.NewTracerServer(svc.Tracer).MiddlewareTracerFunc,
		svc.Localizer,
		func(ctx *gin.Context) {
			if svc.CliCtx != nil {
				ctx.Set(controller.DewuReleaseVersion, svc.CliCtx.String("releaseVersion"))
			}
		})

	svc.Engine.GET(defaultHealthPath, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, &controller.Response{
			Code:   http.StatusOK,
			Status: http.StatusOK,
			Msg:    "health ok",
		})
	})

	return svc
}

func (svc *GinServer) Localizer(ctx *gin.Context) {
	localizer := i18n.NewLocalizer(svc.I18nBundle, ctx.Request.FormValue("lang"), ctx.GetHeader("Accept-Language"), "zh-CN")

	ctx.Set(Localizer, localizer)
	ctx.Set(Language, ctx.GetHeader("Accept-Language"))
	ctx.Next()
}

func (svc *GinServer) String() string {
	return strings.TrimPrefix(svc.ServerCfg.ContextPath, "/")
}
