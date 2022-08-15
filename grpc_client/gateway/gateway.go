package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yousinn/chat-pbx/gateway"
	"github.com/yousinn/go-common/grpc_client"
	"github.com/yousinn/logs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	// grpc options
	grpcKeepAliveTime    = time.Second * 10
	grpcKeepAliveTimeout = time.Second * 3
)

var gatewayGrpcClient *GatewayGrpcClient

type GatewayGrpcClient struct {
	gatewayClientMutx sync.RWMutex
	timeOut           time.Duration
	gatewayClient     gateway.GatewayClient
	gatewayClients    map[string]gateway.GatewayClient
}

func InitGateClients(timeout time.Duration, conn *grpc.ClientConn) {
	gatewayClient := gateway.NewGatewayClient(conn)
	gatewayGrpcClient = &GatewayGrpcClient{
		gatewayClient:  gatewayClient,
		timeOut:        timeout,
		gatewayClients: make(map[string]gateway.GatewayClient),
	}
}

func (s *GatewayGrpcClient) NewGroupClient(addr string) (gateway.GatewayClient, error) {
	ctx1, cancel := context.WithTimeout(context.Background(), s.timeOut)
	defer cancel()
	opts := grpc_client.ClientOpts()
	opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())
	conn, err := grpc.DialContext(ctx1, addr, opts...)
	if err != nil {
		return nil, err
	}

	client := gateway.NewGatewayClient(conn)

	return client, nil
}

func (s *GatewayGrpcClient) GetGatewayGrpcClient(addr string) (gateway.GatewayClient, error) {
	var err error
	s.gatewayClientMutx.RLock()
	gClient, ok := s.gatewayClients[addr]
	s.gatewayClientMutx.RUnlock()
	if !ok {
		s.gatewayClientMutx.Lock()
		defer s.gatewayClientMutx.Unlock()
		gClient, ok = s.gatewayClients[addr]
		if !ok {
			gClient, err = s.NewGroupClient(addr)
			if err != nil {
				return nil, err
			}
			s.gatewayClients[addr] = gClient
		}
	}
	return gClient, nil
}

// 往网关发送消息
func SendMsg(ctx context.Context, addr, id string, uid, sid, scheme string, cid int64, data []byte) error {
	message := &gateway.Content{
		Id:     id,
		Route:  "im.chat",
		Target: &gateway.Target{Uid: uid, Sid: sid, Scheme: scheme, Cid: cid},
		Data:   data,
		Kind:   "response",
	}

	var client gateway.GatewayClient
	var err error
	if addr == "" {
		client = gatewayGrpcClient.gatewayClient
	} else {
		client, err = gatewayGrpcClient.GetGatewayGrpcClient(addr)
		if err != nil {
			logs.Logger.Error("gatewayGrpcClient.GetGatewayGrpcClient error", zap.String("addr", addr))
			client = gatewayGrpcClient.gatewayClient
		}
	}

	if client == nil {
		logs.Logger.Error("gatewayClient is nil")
		return fmt.Errorf("gatewayClient is nil")
	}
	ctx2, cancel := context.WithTimeout(ctx, gatewayGrpcClient.timeOut)
	defer cancel()
	reply, err := client.Send(ctx2, message)
	if err != nil {
		return fmt.Errorf("gatewayClient send failed. err[%v] msg[%+v]", err, message)
	}
	items := reply.GetItems()
	if len(items) == 0 || items[0].GetReply().GetCode() != 0 {
		if items[0].GetReply().GetCode() == 3 {
			logs.Logger.Info("gatewayClient sendmsg fail user not online", zap.Any("reply ", reply), zap.Any("message ", message))
			return nil
		}
		return fmt.Errorf("gatewayClient send reply error. reply[%+v] msg[%+v]", reply, message)
	}

	logs.Logger.Debug("gatewayClient Send success", zap.Any("reply", reply), zap.Any("content", message))
	return nil
}

// SendMsgToGateway live send msg to gateway
func SendMsgToGateway(ctx context.Context, addr string, message *gateway.Content) error {
	var (
		client gateway.GatewayClient
		err    error
	)
	if addr == "" {
		client = gatewayGrpcClient.gatewayClient
	} else {
		client, err = gatewayGrpcClient.GetGatewayGrpcClient(addr)
		if err != nil {
			logs.Logger.Error("gatewayGrpcClient.GetGatewayGrpcClient error", zap.String("addr", addr))
			client = gatewayGrpcClient.gatewayClient
		}
	}

	if client == nil {
		logs.Logger.Error("gatewayClient is nil")
		return fmt.Errorf("gatewayClient is nil")
	}
	ctx2, cancel := context.WithTimeout(ctx, gatewayGrpcClient.timeOut)
	defer cancel()
	reply, err := client.Send(ctx2, message)
	if err != nil {
		return fmt.Errorf("gatewayClient send failed. err[%v] msg[%+v]", err, message)
	}
	items := reply.GetItems()
	if len(items) == 0 || items[0].GetReply().GetCode() != 0 {
		if items[0].GetReply().GetCode() == 3 {
			logs.Logger.Info("gatewayClient sendmsg fail user not online", zap.Any("reply ", reply), zap.Any("message ", message))
			return nil
		}
		return fmt.Errorf("gatewayClient send reply error. reply[%+v] msg[%+v]", reply, message)
	}
	return nil
}

func KickGatewaySession(ctx context.Context, addr string, uid, sid, scheme string, cid int64, group string) error {
	target := &gateway.Target{Uid: uid, Scheme: scheme, Cid: cid, Sid: sid, Group: group}
	var client gateway.GatewayClient
	var err error
	if addr == "" {
		client = gatewayGrpcClient.gatewayClient
	} else {
		client, err = gatewayGrpcClient.GetGatewayGrpcClient(addr)
		if err != nil {
			logs.Logger.Error("gatewayGrpcClient.GetGatewayGrpcClient error", zap.String("addr", addr))
			client = gatewayGrpcClient.gatewayClient
		}
	}
	if client == nil {
		logs.Logger.Error("gatewayClient is nil")
		return fmt.Errorf("gatewayClient is nil")
	}
	ctx2, cancel := context.WithTimeout(ctx, gatewayGrpcClient.timeOut)
	defer cancel()
	res, err := client.Kick(ctx2, target)
	if err != nil {
		return fmt.Errorf("gatewayClient kick failed. err[%v] msg[%+v]", err, target)
	}
	logs.Logger.Debug("gatewayClient Kick success", zap.Any("reply", res), zap.Any("target", target))
	return nil
}
