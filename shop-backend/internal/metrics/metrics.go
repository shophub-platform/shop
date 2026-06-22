package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests partitioned by method, path and status code.",
		},
		[]string{"method", "path", "status_code"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency partitioned by method and path.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// HTTPResponseBytesTotal tracks total bytes sent; divide by 1e9 for GB.
	HTTPResponseBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_bytes_total",
			Help: "Total bytes sent in HTTP responses partitioned by method and path.",
		},
		[]string{"method", "path"},
	)

	// HTTPUniqueVisitorsTotal counts new IP+UA+day combinations seen today.
	HTTPUniqueVisitorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_unique_visitors_total",
			Help: "Unique daily visitors identified by IP and user-agent.",
		},
	)

	OrdersCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders successfully created.",
		},
	)

	// ItemsStockLevel is a per-item gauge updated on create/update/delete.
	ItemsStockLevel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "items_stock_level",
			Help: "Current stock level per item.",
		},
		[]string{"item_id", "item_name"},
	)

	// PaymentProcessingDuration measures time from order creation to confirmation.
	PaymentProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "payment_processing_duration_seconds",
			Help:    "Duration from order creation to payment confirmation in seconds.",
			Buckets: []float64{5, 15, 30, 60, 120, 300, 600},
		},
	)
)

func Register() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPResponseBytesTotal,
		HTTPUniqueVisitorsTotal,
		OrdersCreatedTotal,
		ItemsStockLevel,
		PaymentProcessingDuration,
	)
}
