package session

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopastro/chat-pbx/session"
	"github.com/shopastro/go-common/grpc_client"
	"github.com/shopastro/logs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"strings"
)

const sessionGrpcClientServiceName = "im-session"

var sessionClient session.SessionServiceClient

func SetsessionClient(conn *grpc.ClientConn) {
	sessionClient = session.NewSessionServiceClient(conn)
}

func GetClient() session.SessionServiceClient {
	return sessionClient
}

func GetSessionInfo(ctx context.Context, scheme, dewuUid string, sid string, uid int64) (*session.Session, error) {
	in := &session.User{
		Uid:    dewuUid,
		Scheme: scheme,
		Sid:    sid,
		Cid:    uid,
	}
	if sessionClient == nil {
		logs.Logger.Error("getewaySessionClient is nil")
		return nil, errors.New("getewaySessionClient is nil")
	}
	c, cancel := context.WithTimeout(ctx, grpc_client.GetGrpcClient(sessionGrpcClientServiceName).TimeOut)
	defer cancel()
	sess, err := sessionClient.Find(c, in)
	return sess, err
}

func GetUidByUname(ctx context.Context, uname string) (int64, error) {
	scheme, dewuUid, err := unameToScheme(uname)
	if err != nil {
		return 0, err
	}
	return GetUidByDewuUid(ctx, scheme, dewuUid, "")
}

func schemeToUname(scheme, userId string) string {
	return fmt.Sprintf("%s:%s", scheme, userId)
}

func unameToScheme(uname string) (string, string, error) {
	if strings.Contains(uname, ":") {
		unames := strings.Split(uname, ":")
		return unames[0], unames[1], nil
	} else {
		return "", "", fmt.Errorf("format error")
	}
}

// from:(cache, database)  Where does the data come from?
func GetUidByDewuUid(ctx context.Context, scheme, dewuUid string, from string) (int64, error) {
	if scheme == "" || dewuUid == "" {
		logs.Logger.Error("GetTinodeUid params error", zap.String("scheme", scheme),
			zap.String("dewuUid", dewuUid), zap.String("from", from))
		return 0, fmt.Errorf("params error")
	}
	mapping := &session.Mapping{Scheme: scheme, Uid: dewuUid, From: from}
	res, err := sessionClient.GetMapping(ctx, mapping)
	if err != nil {
		logs.Logger.Error("sessionClient.SetMapping error", zap.Error(err))
		return 0, err
	}

	if !res.Found || res.Mapping == nil {
		logs.Logger.Debug("res Mapping is nil ", zap.String("scheme", scheme), zap.String("dewuUid", dewuUid))
		return 0, nil
	}
	//logs.Logger.Debug("sessionClient.GetMapping res", zap.Any("res", res), zap.Any("mapping", mapping))

	return res.Mapping.Cid, nil
}

func GetDewuUid(ctx context.Context, cid int64) (string, string, error) {
	if cid <= 0 {
		logs.Logger.Error("GetTinodeUid params error")
		return "", "", fmt.Errorf("params error")
	}
	mapping := &session.Mapping{Cid: cid}
	res, err := sessionClient.GetMapping(ctx, mapping)
	if err != nil {
		logs.Logger.Error("sessionClient.SetMapping error", zap.Error(err))
		return "", "", err
	}

	if !res.Found || res.Mapping == nil {
		logs.Logger.Debug("res Mapping is nil ", zap.Int64("cid", cid))
		return "", "", nil
	}

	//logs.Logger.Debug("sessionClient.GetMapping res", zap.Any("res", res), zap.Any("mapping", mapping))
	return res.Mapping.Scheme, res.Mapping.Uid, nil
}

func WithMetaData(ctx context.Context, uid int64, key, value string) {
	//logs.Logger.Debug("WithMetaData", zap.Int64("uid", uid), zap.String("key", key), zap.String("value", value))
	//ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	kv := make(map[string]string)
	//strSeq := strconv.FormatInt(seq, 10)
	kv[key] = value
	in := &session.Metadata{
		User: &session.User{Cid: uid},
		Kv:   kv,
	}
	if sessionClient == nil {
		logs.Logger.Error("sessionClient is nil")
		return
	}
	_, err := sessionClient.WithMetadata(ctx, in)
	if err != nil {
		logs.Logger.Error("gatewayClient.With error", zap.Error(err), zap.Int64("uid", uid))
	}
}

// GetDuUidFromMapping : params scheme uid
func GetDuUidFromMapping(ctx context.Context, mapping *session.Mapping) (string, string, error) {
	if mapping.Cid <= 0 {
		logs.Logger.Error("GetTinodeUid params error")
		return "", "", fmt.Errorf("params error")
	}
	//mapping := &session.Mapping{Cid: cid}
	res, err := sessionClient.GetMapping(ctx, mapping)
	if err != nil {
		logs.Logger.Error("sessionClient.SetMapping error", zap.Error(err))
		return "", "", err
	}

	if !res.Found || res.Mapping == nil {
		logs.Logger.Debug("res Mapping is nil ", zap.Int64("cid", mapping.Cid))
		return "", "", nil
	}
	return res.Mapping.Scheme, res.Mapping.Uid, nil
}

// GetUidFromMapping from:(cache, database)  Where does the data come from?
func GetUidFromMapping(ctx context.Context, mapping *session.Mapping) (int64, error) {
	if mapping.Scheme == "" || mapping.Uid == "" {
		logs.Logger.Error("GetTinodeUid params error", zap.Any("mapping", mapping))
		return 0, fmt.Errorf("params error")
	}
	res, err := sessionClient.GetMapping(ctx, mapping)
	if err != nil {
		logs.Logger.Error("sessionClient.SetMapping error", zap.Error(err))
		return 0, err
	}

	if !res.Found || res.Mapping == nil {
		logs.Logger.Debug("res Mapping is nil ", zap.Any("mapping", mapping))
		return 0, nil
	}
	return res.Mapping.Cid, nil
}
