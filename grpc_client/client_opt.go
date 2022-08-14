package grpc_client

import (
	"encoding/json"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

func ClientOpts() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			grpc_prometheus.UnaryClientInterceptor,
		),
		grpc.WithChainStreamInterceptor(
			grpc_prometheus.StreamClientInterceptor,
		),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithInitialWindowSize(grpcInitialWindowSize),
		grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
		grpc.WithDefaultServiceConfig(getGrpcRoundrobin()),
	}
}
func getGrpcRoundrobin() string {
	var loadBalancer = struct {
		LoadBalancingPolicy string `json:"loadBalancingPolicy"`
	}{
		LoadBalancingPolicy: roundrobin.Name,
	}

	body, err := json.Marshal(loadBalancer)
	if err != nil {
		return roundrobin.Name
	}

	return string(body)
}
