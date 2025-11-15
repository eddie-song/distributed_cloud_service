package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// KV store metrics
	KVPutOperations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kv_put_operations_total",
			Help: "Total number of PUT operations",
		},
	)

	KVGetOperations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kv_get_operations_total",
			Help: "Total number of GET operations",
		},
	)

	KVDeleteOperations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kv_delete_operations_total",
			Help: "Total number of DELETE operations",
		},
	)

	KVStoreSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kv_store_size",
			Help: "Current number of keys in the store",
		},
	)

	// Raft metrics
	RaftIsLeader = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "raft_is_leader",
			Help: "1 if this node is the leader, 0 otherwise",
		},
	)

	RaftAppliedIndex = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "raft_applied_index",
			Help: "Last applied Raft log index",
		},
	)

	RaftCommitIndex = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "raft_commit_index",
			Help: "Last committed Raft log index",
		},
	)
)

