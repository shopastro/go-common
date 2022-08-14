package grpc_client

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/shopastro/logs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	GrpcClientServiceEtcd struct {
		router map[string]string
		conn   map[string]*grpc.ClientConn
	}
)

var (
	conn *GrpcClientServiceEtcd
	err  error
)

func GetGrpcClientEtcd(key string) *grpc.ClientConn {
	c, ok := conn.conn[key]
	if !ok {
		return nil
	}

	return c
}

func NewGrpcClientServiceEtcd(router map[string]string) *GrpcClientServiceEtcd {
	conn = &GrpcClientServiceEtcd{
		conn:   make(map[string]*grpc.ClientConn),
		router: router,
	}

	return conn
}

func (svc *GrpcClientServiceEtcd) DialEtcd() *GrpcClientServiceEtcd {
	grpc_prometheus.EnableClientHandlingTimeHistogram()
	for k, v := range svc.router {
		svc.conn[k], err = grpc.Dial(v, grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
				grpc_opentracing.UnaryClientInterceptor(),
				grpc_prometheus.UnaryClientInterceptor,
			)),
			grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
				grpc_opentracing.StreamClientInterceptor(),
				grpc_prometheus.StreamClientInterceptor,
			)),
			grpc.WithKeepaliveParams(kacp),
			grpc.WithBackoffMaxDelay(BackoffMaxDelay),
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			grpc.WithDefaultServiceConfig(getGrpcRoundrobin()),
		)
		if err != nil {
			logs.Logger.Error("[GrpcClient Dial]", zap.Error(err))
		}
	}

	return svc
}
