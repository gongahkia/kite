package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Scraping metrics
	ScrapingTotal        *prometheus.CounterVec
	ScrapingDuration     *prometheus.HistogramVec
	ScrapingErrors       *prometheus.CounterVec
	CasesScraped         *prometheus.CounterVec
	ScrapingQueueDepth   prometheus.Gauge

	// Worker metrics
	WorkerUtilization    prometheus.Gauge
	WorkerJobsProcessed  *prometheus.CounterVec
	WorkerJobDuration    *prometheus.HistogramVec
	WorkerJobErrors      *prometheus.CounterVec

	// Queue metrics
	QueueDepth           *prometheus.GaugeVec
	QueueEnqueueTotal    *prometheus.CounterVec
	QueueDequeueTotal    *prometheus.CounterVec
	QueueProcessingTime  *prometheus.HistogramVec

	// Storage metrics
	StorageOperations    *prometheus.CounterVec
	StorageErrors        *prometheus.CounterVec
	StorageLatency       *prometheus.HistogramVec

	// Cache metrics
	CacheHits            *prometheus.CounterVec
	CacheMisses          *prometheus.CounterVec
	CacheSize            prometheus.Gauge

	// Citation metrics
	CitationsExtracted   *prometheus.CounterVec
	CitationNetworkNodes prometheus.Gauge
	CitationNetworkEdges prometheus.Gauge

	// Validation metrics
	ValidationTotal      *prometheus.CounterVec
	ValidationErrors     *prometheus.CounterVec
	QualityScores        *prometheus.HistogramVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kite_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kite_http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// Scraping metrics
		ScrapingTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_scraping_total",
				Help: "Total number of scraping operations",
			},
			[]string{"jurisdiction", "source", "status"},
		),
		ScrapingDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kite_scraping_duration_seconds",
				Help:    "Scraping operation duration in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
			},
			[]string{"jurisdiction", "source"},
		),
		ScrapingErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_scraping_errors_total",
				Help: "Total number of scraping errors",
			},
			[]string{"jurisdiction", "source", "error_type"},
		),
		CasesScraped: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_cases_scraped_total",
				Help: "Total number of cases scraped",
			},
			[]string{"jurisdiction", "source"},
		),
		ScrapingQueueDepth: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kite_scraping_queue_depth",
				Help: "Current depth of scraping queue",
			},
		),

		// Worker metrics
		WorkerUtilization: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kite_worker_utilization",
				Help: "Worker pool utilization (0-1)",
			},
		),
		WorkerJobsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_worker_jobs_processed_total",
				Help: "Total number of jobs processed by workers",
			},
			[]string{"worker_id", "job_type", "status"},
		),
		WorkerJobDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kite_worker_job_duration_seconds",
				Help:    "Worker job duration in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
			},
			[]string{"job_type"},
		),
		WorkerJobErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_worker_job_errors_total",
				Help: "Total number of worker job errors",
			},
			[]string{"job_type", "error_type"},
		),

		// Queue metrics
		QueueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kite_queue_depth",
				Help: "Current queue depth",
			},
			[]string{"queue_name"},
		),
		QueueEnqueueTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_queue_enqueue_total",
				Help: "Total number of items enqueued",
			},
			[]string{"queue_name"},
		),
		QueueDequeueTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_queue_dequeue_total",
				Help: "Total number of items dequeued",
			},
			[]string{"queue_name"},
		),
		QueueProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kite_queue_processing_time_seconds",
				Help:    "Queue item processing time in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"queue_name"},
		),

		// Storage metrics
		StorageOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_storage_operations_total",
				Help: "Total number of storage operations",
			},
			[]string{"operation", "status"},
		),
		StorageErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_storage_errors_total",
				Help: "Total number of storage errors",
			},
			[]string{"operation", "error_type"},
		),
		StorageLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kite_storage_latency_seconds",
				Help:    "Storage operation latency in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation"},
		),

		// Cache metrics
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_name"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_name"},
		),
		CacheSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kite_cache_size",
				Help: "Current cache size in bytes",
			},
		),

		// Citation metrics
		CitationsExtracted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_citations_extracted_total",
				Help: "Total number of citations extracted",
			},
			[]string{"jurisdiction", "format"},
		),
		CitationNetworkNodes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kite_citation_network_nodes",
				Help: "Number of nodes in citation network",
			},
		),
		CitationNetworkEdges: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kite_citation_network_edges",
				Help: "Number of edges in citation network",
			},
		),

		// Validation metrics
		ValidationTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_validation_total",
				Help: "Total number of validation operations",
			},
			[]string{"model_type", "status"},
		),
		ValidationErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kite_validation_errors_total",
				Help: "Total number of validation errors",
			},
			[]string{"model_type", "error_type"},
		),
		QualityScores: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kite_quality_scores",
				Help:    "Quality scores distribution",
				Buckets: []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
			},
			[]string{"model_type"},
		),
	}

	return m
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordScraping records a scraping operation metric
func (m *Metrics) RecordScraping(jurisdiction, source, status string, duration time.Duration, casesCount int) {
	m.ScrapingTotal.WithLabelValues(jurisdiction, source, status).Inc()
	m.ScrapingDuration.WithLabelValues(jurisdiction, source).Observe(duration.Seconds())
	if casesCount > 0 {
		m.CasesScraped.WithLabelValues(jurisdiction, source).Add(float64(casesCount))
	}
}

// RecordScrapingError records a scraping error
func (m *Metrics) RecordScrapingError(jurisdiction, source, errorType string) {
	m.ScrapingErrors.WithLabelValues(jurisdiction, source, errorType).Inc()
}

// RecordWorkerJob records a worker job metric
func (m *Metrics) RecordWorkerJob(workerID string, jobType, status string, duration time.Duration) {
	m.WorkerJobsProcessed.WithLabelValues(workerID, jobType, status).Inc()
	m.WorkerJobDuration.WithLabelValues(jobType).Observe(duration.Seconds())
}

// Handler returns the Prometheus metrics HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}
