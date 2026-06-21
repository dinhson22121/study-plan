package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reg      *prometheus.Registry
	reqTotal *prometheus.CounterVec
	reqDur   *prometheus.HistogramVec
	failures *prometheus.CounterVec
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	reqTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "edu", Subsystem: "http", Name: "requests_total",
		Help: "Total HTTP requests by method, route and status code.",
	}, []string{"method", "route", "status"})
	reqDur := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "edu", Subsystem: "http", Name: "request_duration_seconds",
		Help:    "HTTP request latency in seconds by method, route and status code.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status"})
	failures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "edu", Subsystem: "app", Name: "failures_total",
		Help: "Business operation failures by domain and kind (e.g. parse, upload, notification).",
	}, []string{"domain", "kind"})

	reg.MustRegister(reqTotal, reqDur, failures,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return &Metrics{reg: reg, reqTotal: reqTotal, reqDur: reqDur, failures: failures}
}

// Middleware records request count and latency for every request, labelled by
// the matched route pattern (e.g. /api/v1/quizzes/:id) to keep cardinality bounded.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		m.reqTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		m.reqDur.WithLabelValues(c.Request.Method, route, status).Observe(time.Since(start).Seconds())
	}
}

// Handler serves the Prometheus exposition format for this registry.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{})
}

// RecordFailure increments the business-failure counter. Safe for concurrent use.
func (m *Metrics) RecordFailure(domain, kind string) {
	m.failures.WithLabelValues(domain, kind).Inc()
}
