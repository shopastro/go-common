package server

import (
	"context"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerRecovery() grpc.UnaryServerInterceptor {
	return grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(func(ctx context.Context, p interface{}) (err error) {
		LogRecoverStack(p)
		err = status.Errorf(codes.Internal, "%s", p)
		return err
	}))
}
