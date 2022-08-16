package client

import (
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"time"
)

type (
	Client struct {
		request *request
	}

	Config struct {
		Debug         bool                `json:"debug" yaml:"debug"`
		Remotes       map[string][]string `json:"remotes" yaml:"remotes"`
		Resp          IResponse
		Tracer        opentracing.Tracer
		Timeout       map[string]map[string]int `json:"httpTimeout"`
		EnableMetrics bool                      `json:"enableMetrics" yaml:"enableMetrics"`
	}
)

var (
	config *Config
)

// init client
func NewClient(cfg *Config) {
	config = cfg
}

// get client
func GetClient(ctxs ...context.Context) *Client {
	var ctx context.Context

	if len(ctxs) == 1 {
		ctx = ctxs[0]
	} else {
		ctx = context.Background()
	}

	return &Client{
		request: NewRequest(config.Remotes, ctx, config.Tracer, config.Debug, config.Timeout, config.EnableMetrics),
	}
}

//request get
func (c *Client) Get(
	remote,
	path string,
	queryParams interface{}) IResponse {

	return c.request.SetRemote(remote).SetPath(path).SetParam(queryParams).SetTimeOutConfig(remote, path).Get()
}

//request post
func (c *Client) Post(
	remote,
	path string,
	dataForm interface{}) IResponse {

	return c.request.SetRemote(remote).SetPath(path).SetParam(dataForm).SetTimeOutConfig(remote, path).Post()
}

//request post
func (c *Client) PostUrlEncode(
	remote,
	path string,
	dataForm interface{}) IResponse {

	return c.request.SetRemote(remote).SetPath(path).SetParam(dataForm).SetTimeOutConfig(remote, path).PostUrlEncode()
}

//request put
func (c *Client) Put(
	remote,
	path string,
	dataForm interface{}) IResponse {

	return c.request.SetRemote(remote).SetPath(path).SetParam(dataForm).SetTimeOutConfig(remote, path).Put()
}

//request post json
func (c *Client) PostJson(
	remote,
	path string,
	dataJson interface{}) IResponse {

	return c.request.SetRemote(remote).SetPath(path).SetParam(dataJson).SetTimeOutConfig(remote, path).PostJson()
}

// request delete
func (c *Client) Delete(
	remote,
	path string,
	dataForm interface{}) IResponse {

	return c.request.SetRemote(remote).SetPath(path).SetParam(dataForm).SetTimeOutConfig(remote, path).Delete()
}

// set request header data params
func (c *Client) SetHeader(data map[string][]string) *Client {
	c.request.Header = data

	return c
}

// set request time out params
func (c *Client) SetTimeOut(times time.Duration) *Client {
	c.request.SetTimeOut(times)

	return c
}

// add request time out params
func (c *Client) AddParams(key, value string) {
	c.request.SuperAgent.QueryData.Add(key, value)
}

func (c *Client) InsecureSkipVerify(skipVerify bool) *Client {
	c.request.InsecureSkipVerify = skipVerify

	return c
}
