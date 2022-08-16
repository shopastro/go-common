package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/yousinn/go-common/common"
	"github.com/yousinn/go-common/gorequest"
	"github.com/yousinn/logs"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"strings"
	"time"
)

type (
	request struct {
		remote             string
		url                string
		SuperAgent         *gorequest.SuperAgent
		remotes            map[string][]string
		timeOut            time.Duration
		timeoutConfig      map[string]map[string]int
		param              interface{}
		Header             map[string][]string
		InsecureSkipVerify bool
		ctx                context.Context
		Response           IResponse
		enableMetrics      bool
	}
)

var (
	TimeOut     = 3 * time.Second
	DefualtJson = `{"status": 900,"msg": "失败","data": ""}`
)

func NewRequest(remote map[string][]string, ctx context.Context, trace opentracing.Tracer, debug bool, timeoutConfig map[string]map[string]int, enableMetrics bool) *request {
	return &request{
		ctx:           ctx,
		timeOut:       TimeOut,
		remotes:       remote,
		Header:        make(map[string][]string),
		Response:      NewResponse(ctx),
		SuperAgent:    gorequest.New().SetGinContext(ctx).SetTrace(trace).SetDebug(debug),
		timeoutConfig: timeoutConfig,
		enableMetrics: enableMetrics,
	}
}

func (svc *request) insecureSkipVerify() {
	if svc.InsecureSkipVerify {
		svc.SuperAgent.TLSClientConfig(
			&tls.Config{
				InsecureSkipVerify: svc.InsecureSkipVerify,
			})
	}
}

// set request time out config
func (svc *request) SetTimeOutConfig(remote, path string) *request {
	rc, ok := svc.timeoutConfig[remote]
	if !ok {
		return svc
	}
	if pt, ok := rc[strings.ToLower(path)]; ok {
		svc.timeOut = time.Duration(pt) * time.Millisecond
		return svc
	}
	if dt, ok := rc["default"]; ok {
		svc.timeOut = time.Duration(dt) * time.Millisecond
		return svc
	}

	return svc
}

func (svc *request) SetTimeOut(times time.Duration) {
	svc.timeOut = times
}

func (svc *request) SetRemote(remote string) *request {
	remoteArray, ok := svc.remotes[remote]
	if ok {
		max := len(remoteArray)
		num := common.NewTools().GenerateRangeNum(0, max)
		svc.remote = remoteArray[num]
	}

	return svc
}

func (svc *request) SetPath(path string) *request {
	path = strings.TrimRight(strings.TrimLeft(path, "/"), "/")

	svc.url = fmt.Sprintf("%s/%s",
		strings.Trim(svc.remote, "/"),
		path)

	svc.url = strings.Trim(svc.url, "/")

	return svc
}

func (svc *request) SetParam(param interface{}) *request {
	svc.param = param

	return svc
}

func (svc *request) Get() IResponse {
	svc.SuperAgent.Timeout(svc.timeOut).Get(svc.url)
	svc.SuperAgent.Header = svc.Header
	svc.insecureSkipVerify()

	res, body, err := svc.SuperAgent.Query(svc.param).End()
	if err == nil && res != nil {
		return svc.Response.SetBody(body, res.StatusCode)

	}
	logs.Logger.Error("[Get]", zap.Any("err", err), zap.String("remote", svc.remote), zap.Any("param", svc.param))
	return svc.Response.SetBody(DefualtJson)
}

func (svc *request) Post() IResponse {
	svc.SuperAgent.Timeout(svc.timeOut).Post(svc.url).Send(svc.param)
	svc.SuperAgent.Header = svc.Header
	svc.insecureSkipVerify()

	res, body, err := svc.SuperAgent.End()
	if err == nil && res != nil {
		return svc.Response.SetBody(body, res.StatusCode)
	}
	logs.Logger.Error("[Post]", zap.Any("err", err), zap.String("remote", svc.remote), zap.Any("param", svc.param))
	return svc.Response.SetBody(DefualtJson)
}

func (svc *request) PostUrlEncode() IResponse {
	svc.SuperAgent.Timeout(svc.timeOut).Post(svc.url).Send(svc.param)
	svc.SuperAgent.Header = svc.Header
	svc.insecureSkipVerify()

	res, body, err := svc.SuperAgent.End()
	if err == nil && res != nil {
		return svc.Response.SetBody(body, res.StatusCode)
	}
	logs.Logger.Error("[PostUrlEncode]", zap.Any("err", err), zap.String("remote", svc.remote), zap.Any("param", svc.param))
	return svc.Response.SetBody(DefualtJson)
}

func (svc *request) Put() IResponse {
	svc.SuperAgent.Timeout(svc.timeOut).Put(svc.url).Send(svc.param)
	svc.SuperAgent.Header = svc.Header
	svc.insecureSkipVerify()

	res, body, err := svc.SuperAgent.End()
	if err == nil && res != nil {
		return svc.Response.SetBody(body, res.StatusCode)
	}
	logs.Logger.Error("[Put]", zap.Any("err", err), zap.String("remote", svc.remote), zap.Any("param", svc.param))
	return svc.Response.SetBody(DefualtJson)
}

func (svc *request) PostJson() IResponse {
	svc.SuperAgent.Timeout(svc.timeOut).Post(svc.url)
	svc.SuperAgent.Header = svc.Header
	svc.insecureSkipVerify()

	paramsJson, errMsg := json.Marshal(svc.param)
	if errMsg != nil {
		return svc.Response.SetBody(DefualtJson, 406)
	}

	res, body, err := svc.SuperAgent.Send(string(paramsJson)).End()
	if err == nil && res != nil {
		return svc.Response.SetBody(body, res.StatusCode)
	}
	logs.Logger.Error("[PostJson]", zap.Any("err", err), zap.String("remote", svc.remote), zap.Any("param", svc.param))
	return svc.Response.SetBody(DefualtJson)
}

func (svc *request) Delete() IResponse {
	svc.SuperAgent.Timeout(svc.timeOut).Delete(svc.url).Send(svc.param)
	svc.SuperAgent.Header = svc.Header
	svc.insecureSkipVerify()

	res, body, err := svc.SuperAgent.End()
	if err == nil && res != nil {
		return svc.Response.SetBody(body, res.StatusCode)
	}
	logs.Logger.Error("[Delete]", zap.Any("err", err), zap.String("remote", svc.remote), zap.Any("param", svc.param))
	return svc.Response.SetBody(DefualtJson)
}
