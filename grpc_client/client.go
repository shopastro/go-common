package grpc_client

import (
	"crypto/tls"
	"fmt"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/shopastro/logs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type (
	GrpcClientService struct {
		router     map[string]*GrpcConfig
		clientConn map[string]*ClientConn
	}

	GrpcConfig struct {
		Addr    string
		Port    int
		TimeOut time.Duration
		Tls     bool
	}

	ClientConn struct {
		Client  *grpc.ClientConn
		TimeOut time.Duration
	}
)

const (
	grpcInitialWindowSize     = 2 << 24
	grpcInitialConnWindowSize = 2 << 24
	grpcMaxSendMsgSize        = 2 << 24
	grpcMaxCallMsgSize        = 2 << 24
	grpcKeepAliveTime         = 10 * time.Second
	grpcKeepAliveTimeout      = 3 * time.Second
	BackoffMaxDelay           = 3 * time.Second
)

var (
	clientSvr *GrpcClientService
	kacp      = keepalive.ClientParameters{
		Time:                grpcKeepAliveTime,
		Timeout:             grpcKeepAliveTimeout,
		PermitWithoutStream: true,
	}
)

func GetGrpcClient(key string) *ClientConn {
	c, ok := clientSvr.clientConn[key]
	if !ok {
		return nil
	}

	return c
}

func NewGrpcClientService(router map[string]*GrpcConfig) *GrpcClientService {
	clientSvr = &GrpcClientService{
		clientConn: make(map[string]*ClientConn),
		router:     router,
	}
	return clientSvr
}

func (svc *GrpcClientService) Dial() *GrpcClientService {
	var secOpt grpc.DialOption
	grpc_prometheus.EnableClientHandlingTimeHistogram()
	for name, cfg := range svc.router {
		if cfg.Tls {
			secOpt = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: cfg.Tls,
			}))
		} else {
			secOpt = grpc.WithInsecure()
		}
		clientConn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Addr, cfg.Port), append(ClientOpts(), secOpt)...)
		if err != nil {
			logs.Logger.Error("[GrpcClient Dial]", zap.Error(err))
			continue
		}

		svc.clientConn[name] = &ClientConn{
			Client:  clientConn,
			TimeOut: cfg.TimeOut * time.Millisecond,
		}
	}
	return svc
}
