package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Нужно создать свой Registry чтобы не получать ненужных метрик
	Registry = prometheus.NewRegistry()

	RequestsTotal = promauto.With(Registry).NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	ResponseTime = promauto.With(Registry).NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_time_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{0.1, 0.5, 1, 2, 5},
	}, []string{"method", "path"})

	PVZCreated = promauto.With(Registry).NewCounter(prometheus.CounterOpts{
		Name: "business_pvz_created_total",
		Help: "Total number of created PVZs",
	})

	ReceptionsCreated = promauto.With(Registry).NewCounter(prometheus.CounterOpts{
		Name: "business_receptions_created_total",
		Help: "Total number of created receptions",
	})

	ProductsAdded = promauto.With(Registry).NewCounter(prometheus.CounterOpts{
		Name: "business_products_added_total",
		Help: "Total number of added products",
	})
)
