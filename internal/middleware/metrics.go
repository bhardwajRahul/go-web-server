package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics.
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// Database metrics.
	databaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	databaseConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	databaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"operation", "table"},
	)

	databaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	// Application metrics.
	applicationInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "application_info",
			Help: "Application information",
		},
		[]string{"version", "go_version", "environment"},
	)

	applicationStartTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "application_start_time_seconds",
			Help: "Unix timestamp of when the application started",
		},
	)

	// HTMX specific metrics.
	htmxRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "htmx_requests_total",
			Help: "Total number of HTMX requests",
		},
		[]string{"trigger", "target", "swap"},
	)

	// CSRF metrics.
	csrfTokensGenerated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "csrf_tokens_generated_total",
			Help: "Total number of CSRF tokens generated",
		},
	)

	csrfValidationFailures = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "csrf_validation_failures_total",
			Help: "Total number of CSRF validation failures",
		},
	)

	// User activity metrics.
	usersCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_created_total",
			Help: "Total number of users created",
		},
	)

	usersActiveTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "users_active_total",
			Help: "Total number of active users",
		},
	)
)

// PrometheusMiddleware creates HTTP metrics middleware.
func PrometheusMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Increment in-flight requests
			httpRequestsInFlight.Inc()
			defer httpRequestsInFlight.Dec()

			// Process request
			err := next(c)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(c.Response().Status)
			method := c.Request().Method
			path := c.Path()

			httpRequestsTotal.WithLabelValues(method, path, status).Inc()
			httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)

			// Track HTMX requests
			if c.Request().Header.Get("HX-Request") == "true" {
				trigger := c.Request().Header.Get("HX-Trigger")
				target := c.Request().Header.Get("HX-Target")
				swap := c.Request().Header.Get("HX-Swap")
				htmxRequestsTotal.WithLabelValues(trigger, target, swap).Inc()
			}

			return err
		}
	}
}

// InitializeMetrics initializes application metrics with static information.
func InitializeMetrics(version, goVersion, environment string) {
	applicationInfo.WithLabelValues(version, goVersion, environment).Set(1)
	applicationStartTime.Set(float64(time.Now().Unix()))
}

// UpdateDatabaseMetrics updates database connection metrics.
func UpdateDatabaseMetrics(active, idle int) {
	databaseConnectionsActive.Set(float64(active))
	databaseConnectionsIdle.Set(float64(idle))
}

// RecordDatabaseQuery records database query metrics.
func RecordDatabaseQuery(operation, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	databaseQueriesTotal.WithLabelValues(operation, table, status).Inc()
	databaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCSRFTokenGenerated increments CSRF token generation counter.
func RecordCSRFTokenGenerated() {
	csrfTokensGenerated.Inc()
}

// RecordCSRFValidationFailure increments CSRF validation failure counter.
func RecordCSRFValidationFailure() {
	csrfValidationFailures.Inc()
}

// RecordUserCreated increments user creation counter.
func RecordUserCreated() {
	usersCreatedTotal.Inc()
}

// UpdateActiveUsers updates the active users gauge.
func UpdateActiveUsers(count int64) {
	usersActiveTotal.Set(float64(count))
}
