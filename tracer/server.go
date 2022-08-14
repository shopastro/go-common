package tracer

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type (
	TracerServer struct {
		options []MWOption
		tracer  opentracing.Tracer
	}

	mwOptions struct {
		opNameFunc    func(r *http.Request) string
		spanFilter    func(r *http.Request) bool
		spanObserver  func(span opentracing.Span, r *http.Request)
		urlTagFunc    func(u *url.URL) string
		componentName string
	}

	MWOption func(*mwOptions)
)

const (
	defaultComponentName = "net/http"
)

func NewTracerServer(tr opentracing.Tracer, opts ...MWOption) *TracerServer {
	return &TracerServer{
		tracer:  tr,
		options: opts,
	}
}

func (svc *TracerServer) MiddlewareTracerFunc(context *gin.Context) {
	opts := mwOptions{
		opNameFunc: func(r *http.Request) string {
			return "SERVER " + r.Method + ": " + r.URL.Path
		},
		spanFilter:   func(r *http.Request) bool { return true },
		spanObserver: func(span opentracing.Span, r *http.Request) {},
		urlTagFunc: func(u *url.URL) string {
			return u.String()
		},
	}

	for _, opt := range svc.options {
		opt(&opts)
	}

	if !opts.spanFilter(context.Request) {
		context.Next()

		return
	}

	ctx, _ := svc.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(context.Request.Header))
	sp := svc.tracer.StartSpan(opts.opNameFunc(context.Request), ext.RPCServerOption(ctx))

	ext.HTTPMethod.Set(sp, context.Request.Method)
	ext.HTTPUrl.Set(sp, opts.urlTagFunc(context.Request.URL))

	opts.spanObserver(sp, context.Request)

	componentName := opts.componentName
	if componentName == "" {
		componentName = defaultComponentName
	}

	ext.Component.Set(sp, componentName)

	context.Request = context.Request.WithContext(opentracing.ContextWithSpan(context.Request.Context(), sp))

	defer func() {
		ext.HTTPStatusCode.Set(sp, uint16(context.Writer.Status()))
		if context.Writer.Status() >= http.StatusInternalServerError {
			ext.Error.Set(sp, true)
		}

		sp.Finish()
	}()

	context.Next()
}
