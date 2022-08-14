package server

import (
	"time"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func ServerOpts(kaEnabled bool) []grpc.ServerOption {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			UnaryServerRecovery(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		),

		grpc.ChainStreamInterceptor(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
		),
	}
	if kaEnabled {
		kepConfig := keepalive.EnforcementPolicy{
			MinTime:             1 * time.Second, // If a client pings more than once every second, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}
		opts = append(opts, grpc.KeepaliveEnforcementPolicy(kepConfig))

		kpConfig := keepalive.ServerParameters{
			Time:    60 * time.Second, // Ping the client if it is idle for 60 seconds to ensure the connection is still active
			Timeout: 20 * time.Second, // Wait 20 second for the ping ack before assuming the connection is dead
		}

		opts = append(opts, grpc.KeepaliveParams(kpConfig))
	}

	return opts
}
