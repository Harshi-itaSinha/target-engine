package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	CampaignsMatched *prometheus.HistogramVec
	ActiveCampaigns  prometheus.Gauge
	TargetingRules   prometheus.Gauge
}

func NewMetrics() *Metrics {
	metrics := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "targeting_engine_requests_total",
				Help: "Total number of requests processed",
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "targeting_engine_request_duration_seconds",
				Help:    "Request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		CampaignsMatched: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "targeting_engine_campaigns_matched",
				Help:    "Number of campaigns matched per request",
				Buckets: []float64{0, 1, 2, 5, 10, 20, 50},
			},
			[]string{"country", "os"},
		),
		ActiveCampaigns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "targeting_engine_active_campaigns",
				Help: "Number of active campaigns",
			},
		),
		TargetingRules: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "targeting_engine_targeting_rules",
				Help: "Number of targeting rules",
			},
		),
		
	}

	prometheus.MustRegister(
		metrics.RequestsTotal,
		metrics.RequestDuration,
		metrics.CampaignsMatched,
		metrics.ActiveCampaigns,
		metrics.TargetingRules,
	)

	return metrics
}

func (m *Metrics) RecordRequest(method, endpoint string, statusCode int, duration time.Duration) {
	status := strconv.Itoa(statusCode)
	m.RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	m.RequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

func (m *Metrics) RecordCampaignsMatched(country, os string, count int) {
	m.CampaignsMatched.WithLabelValues(country, os).Observe(float64(count))
}

func (m *Metrics) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		m.RecordRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *metricsResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
