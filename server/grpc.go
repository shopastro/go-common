package server

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type (
	GrpcServer struct {
		Server            *grpc.Server
		Listener          net.Listener
		RegisteGrpcServer func(*grpc.Server)
	}
)

func NewGrpcServer() *GrpcServer {
	return &GrpcServer{}
}

func (svc *GrpcServer) RunGrpcServe() error {
	grpc_prometheus.EnableHandlingTimeHistogram()

	svc.Server = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
		)),

		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
		)),
	)

	svc.RegisteGrpcServer(svc.Server)

	reflection.Register(svc.Server)

	grpc_prometheus.Register(svc.Server)
	return svc.Server.Serve(svc.Listener)
}
