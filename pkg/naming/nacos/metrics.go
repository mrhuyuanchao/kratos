package nacos

import "kratos/pkg/stat/metric"

const (
	clientNamespace = "rpcx_client"
)

var (
	_metricClientReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rpcx client requests duration(ms).",
		Labels:    []string{"path", "method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
	_metricClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "rpcx client requests code count.",
		Labels:    []string{"path", "method", "code"},
	})
)
