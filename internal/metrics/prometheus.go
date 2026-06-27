package metrics

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request metrics
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request duration in seconds",
		},
		[]string{"method", "path"},
	)

	// Database metrics
	DatabaseQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"},
	)

	DatabaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "database_query_duration_seconds",
			Help: "Database query duration in seconds",
		},
		[]string{"operation"},
	)

	// Circuit breaker metrics
	CircuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of circuit breaker (0=closed, 1=open, 2=half-open)",
		},
		[]string{"name"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		DatabaseQueriesTotal,
		DatabaseQueryDuration,
		CircuitBreakerState,
	)
}

// MetricsMiddleware returns middleware that records HTTP metrics
func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			method := c.Request().Method

			timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
				HTTPRequestDuration.WithLabelValues(method, path).Observe(v)
			}))

			err := next(c)

			status := c.Response().Status
			HTTPRequestsTotal.WithLabelValues(method, path, string(rune(status))).Add(1)
			timer.ObserveDuration()

			return err
		}
	}
}

// PrometheusHandler returns an HTTP handler for Prometheus scraping
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// InitPrometheus registers all metrics and sets up the endpoint
func InitPrometheus(e *echo.Echo) {
	e.GET("/metrics", func(c echo.Context) error {
		promhttp.Handler().ServeHTTP(c.Response(), c.Request())
		return nil
	})
}
