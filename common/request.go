package common

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"strings"
	"sync"
)

type Request struct {
}

var (
	r       *Request
	rOnce   sync.Once
	request *Request
)

func init() {
	request = NewRequest()
}

/**
 * 返回单例实例
 * @method New
 */
func NewRequest() *Request {
	rOnce.Do(func() { //只执行一次
		r = &Request{}
	})

	return r
}

// 获取设备号
func (r *Request) GetUuid(ctx *gin.Context) string {
	uuid := ctx.GetHeader("DUUUID")

	if uuid == "" {
		uuid = ctx.GetHeader("UUID")
	}
	if uuid == "" {
		uuid = ctx.Query("uuid")
	}
	if uuid == "" {
		uuid = ctx.DefaultPostForm("uuid", "")
	}

	return uuid
}

// 获取版本号
func (r *Request) GetVersion(ctx *gin.Context) string {
	version := ctx.GetHeader("DUV")

	if version == "" {
		version = ctx.Query("v")
	}
	if version == "" {
		version = ctx.DefaultPostForm("v", "")
	}

	return version
}

// 获取平台
func (r *Request) GetPlatform(ctx *gin.Context) string {
	platform := ctx.GetHeader("DUPLATFORM")

	if platform == "" {
		platform = ctx.Query("platform")
	}
	if platform == "" {
		platform = ctx.DefaultPostForm("platform", "")
	}

	return strings.ToLower(platform)
}

// 是否安卓
func (r *Request) IsAndroid(ctx *gin.Context) bool {
	return r.GetPlatform(ctx) == "android"
}

// 是否IOS
func (r *Request) IsIos(ctx *gin.Context) bool {
	return r.GetPlatform(ctx) == "iphone"
}

func (r *Request) TraceId(ctx *gin.Context) string {
	if span := opentracing.SpanFromContext(ctx.Request.Context()); span != nil {
		if sc, ok := span.Context().(jaeger.SpanContext); ok {
			return sc.TraceID().String()
		}
	}

	return NewTools().CreateUUID()
}
