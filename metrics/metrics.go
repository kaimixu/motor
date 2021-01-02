package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ReqCnt *prometheus.CounterVec
	ReqDur *prometheus.HistogramVec
	ReqErr *prometheus.CounterVec

	DefaultPath = "/metrics"
)

func Init(namespace string) {
	ReqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "requests_total",
			Help:      "http client requests count",
		}, []string{"path", "code"})

	ReqDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "duration_ms",
			Buckets:   []float64{0.5, 1, 2, 5, 10},
			Help:      "http client requests duration(ms)",
		}, []string{"path"})

	ReqErr = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "requests_errcode_total",
			Help:      "http client error requests count",
		}, []string{"path", "code"})

	prometheus.MustRegister(ReqCnt, ReqDur, ReqErr)
}
