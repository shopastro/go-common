package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/shopastro/go-common/controller"
	"github.com/shopastro/logs"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	defaultMetricPath = "/metrics"

	reqCnt = &Metric{
		ID:          "reqCnt",
		Name:        "requests_total",
		Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		Type:        "counter_vec",
		Args:        []string{"code", "status", "method", "handler", "host", "url", "version"}}

	reqDur = &Metric{
		ID:          "reqDur",
		Name:        "request_duration_seconds",
		Description: "The HTTP request latencies in seconds.",
		Type:        "histogram_vec",
		Args:        []string{"code", "status", "method", "handler", "host", "url"}}

	resSz = &Metric{
		ID:          "resSz",
		Name:        "response_size_bytes",
		Description: "The HTTP response sizes in bytes.",
		Type:        "summary_vec",
		Args:        []string{"code", "status", "method", "handler", "host", "url"}}

	reqSz = &Metric{
		ID:          "reqSz",
		Name:        "request_size_bytes",
		Description: "The HTTP request sizes in bytes.",
		Type:        "summary_vec",
		Args:        []string{"code", "status", "method", "handler", "host", "url"}}

	standardMetrics = []*Metric{
		reqCnt,
		reqDur,
		resSz,
		reqSz,
	}
)

type RequestCounterURLLabelMappingFn func(c *gin.Context) string

type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
}

type Prometheus struct {
	reqCnt                  *prometheus.CounterVec
	reqSz, resSz            *prometheus.SummaryVec
	reqDur                  *prometheus.HistogramVec
	router                  *gin.Engine
	listenAddress           string
	Ppg                     PrometheusPushGateway
	MetricsList             []*Metric
	MetricsPath             string
	ReqCntURLLabelMappingFn RequestCounterURLLabelMappingFn
	URLLabelFromContext     string
}

type PrometheusPushGateway struct {
	PushIntervalSeconds time.Duration

	PushGatewayURL string

	MetricsURL string

	Job string
}

func NewPrometheus(subsystem string, customMetricsList ...[]*Metric) *Prometheus {
	var metricsList []*Metric

	if len(customMetricsList) > 1 {
		panic("Too many args. NewPrometheus( string, <optional []*Metric> ).")
	} else if len(customMetricsList) == 1 {
		metricsList = customMetricsList[0]
	}

	for _, metric := range standardMetrics {
		metricsList = append(metricsList, metric)
	}

	p := &Prometheus{
		MetricsList: metricsList,
		MetricsPath: defaultMetricPath,
		ReqCntURLLabelMappingFn: func(c *gin.Context) string {
			return c.Request.URL.Path
		},
	}

	p.registerMetrics(subsystem)
	return p
}

func (p *Prometheus) SetPushGateway(pushGatewayURL, metricsURL string, pushIntervalSeconds time.Duration) {
	p.Ppg.PushGatewayURL = pushGatewayURL
	p.Ppg.MetricsURL = metricsURL
	p.Ppg.PushIntervalSeconds = pushIntervalSeconds

	p.startPushTicker()
}

func (p *Prometheus) SetPushGatewayJob(j string) {
	p.Ppg.Job = j
}

func (p *Prometheus) SetListenAddress(address string) {
	p.listenAddress = address
	if p.listenAddress != "" {
		p.router = gin.Default()
	}
}

func (p *Prometheus) SetListenAddressWithRouter(listenAddress string, r *gin.Engine) {
	p.listenAddress = listenAddress
	if len(p.listenAddress) > 0 {
		p.router = r
	}
}

func (p *Prometheus) SetMetricsPath(e *gin.Engine) {
	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, prometheusHandler())
	}
}

func (p *Prometheus) SetMetricsPathWithAuth(e *gin.Engine, accounts gin.Accounts) {
	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
	}

}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go p.router.Run(p.listenAddress)
	}
}

func (p *Prometheus) getMetrics() []byte {
	response, _ := http.Get(p.Ppg.MetricsURL)

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	return body
}

func (p *Prometheus) getPushGatewayURL() string {
	h, _ := os.Hostname()
	if p.Ppg.Job == "" {
		p.Ppg.Job = "gin"
	}

	return p.Ppg.PushGatewayURL + "/metrics/job/" + p.Ppg.Job + "/instance/" + h
}

func (p *Prometheus) sendMetricsToPushGateway(metrics []byte) {
	req, err := http.NewRequest("POST", p.getPushGatewayURL(), bytes.NewBuffer(metrics))
	client := &http.Client{}
	if _, err = client.Do(req); err != nil {
		logs.Logger.Error("Error sending to push gateway", zap.Any("error", err))
	}
}

func (p *Prometheus) startPushTicker() {
	ticker := time.NewTicker(time.Second * p.Ppg.PushIntervalSeconds)
	go func() {
		for range ticker.C {
			p.sendMetricsToPushGateway(p.getMetrics())
		}
	}()
}

func NewMetric(m *Metric, subsystem string) prometheus.Collector {
	var metric prometheus.Collector

	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}

	return metric
}

func (p *Prometheus) registerMetrics(subsystem string) {

	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, subsystem)
		if err := prometheus.Register(metric); err != nil {
			logs.Logger.Error("[Prometheus registered]", zap.String("name", metricDef.Name))
		}

		switch metricDef {
		case reqCnt:
			p.reqCnt = metric.(*prometheus.CounterVec)
		case reqDur:
			p.reqDur = metric.(*prometheus.HistogramVec)
		case resSz:
			p.resSz = metric.(*prometheus.SummaryVec)
		case reqSz:
			p.reqSz = metric.(*prometheus.SummaryVec)
		}

		metricDef.MetricCollector = metric
	}
}

func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc())

	p.SetMetricsPath(e)
}

func (p *Prometheus) UseWithAuth(e *gin.Engine, accounts gin.Accounts) {
	e.Use(p.HandlerFunc())

	p.SetMetricsPathWithAuth(e, accounts)
}

func (p *Prometheus) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == p.MetricsPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		dewuStatus, ok := c.Get(controller.DewuCode)
		if !ok {
			dewuStatus = "unknown"
		}

		dewuCodeString := fmt.Sprintf("%v", dewuStatus)
		code := strconv.Itoa(c.Writer.Status())

		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(c.Writer.Size())

		url := p.ReqCntURLLabelMappingFn(c)
		p.reqDur.WithLabelValues(code, dewuCodeString, c.Request.Method, c.HandlerName(), c.Request.Host, url).Observe(elapsed)
		if len(p.URLLabelFromContext) > 0 {
			u, found := c.Get(p.URLLabelFromContext)
			if !found {
				u = "unknown"
			}
			url = u.(string)
		}

		dewuReleaseVersion, ok := c.Get(controller.DewuReleaseVersion)
		if !ok {
			dewuReleaseVersion = "unknown"
		}

		p.reqCnt.WithLabelValues(code, dewuCodeString, c.Request.Method, c.HandlerName(), c.Request.Host, url, fmt.Sprintf("%v", dewuReleaseVersion)).Inc()
		p.reqSz.WithLabelValues(code, dewuCodeString, c.Request.Method, c.HandlerName(), c.Request.Host, url).Observe(float64(reqSz))
		p.resSz.WithLabelValues(code, dewuCodeString, c.Request.Method, c.HandlerName(), c.Request.Host, url).Observe(resSz)
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}

	s += len(r.Host)
	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}

	return s
}
