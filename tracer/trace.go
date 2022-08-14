package tracer

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"

	"github.com/shopastro/logs"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"go.uber.org/zap"
)

type (
	TraceConfig struct {
		Param       float64
		HostPort    string
		ServiceName string
		LogSpans    bool
	}
)

var (
	tracerClient opentracing.Tracer
)

func GetTracerClient() opentracing.Tracer {
	return tracerClient
}

func NewTracer(cfg *TraceConfig) opentracing.Tracer {
	configEnv, err := config.FromEnv()
	if err != nil {
		logs.Logger.Error("new tracer config", zap.Error(err))
	}

	if cfg.HostPort == "" {
		cfg.HostPort = configEnv.Reporter.LocalAgentHostPort
	}

	traceCfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			SamplingServerURL: configEnv.ServiceName,
		},

		Reporter: &config.ReporterConfig{
			LogSpans:           cfg.LogSpans,
			LocalAgentHostPort: cfg.HostPort,
		},

		ServiceName: cfg.ServiceName,
	}

	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	tracerClient, _, err = traceCfg.NewTracer(
		config.Logger(jaeger.StdLogger),
		config.Injector(opentracing.HTTPHeaders, propagator),
		config.Extractor(opentracing.HTTPHeaders, propagator),
		config.ZipkinSharedRPCSpan(true),
		config.MaxTagValueLength(256),
		config.PoolSpans(true),
	)
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}

	opentracing.SetGlobalTracer(tracerClient)
	return tracerClient
}
