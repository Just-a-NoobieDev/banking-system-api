package lib

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request metrics
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "banking",
		Name:      "requests_total",
		Help:      "Total number of requests by endpoint",
	}, []string{"endpoint", "method", "status"})

	// Transaction metrics
	transactionCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "banking",
		Name:      "transactions_total",
		Help:      "Total number of transactions by type",
	}, []string{"type"})

	transactionAmount = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "banking",
		Name:      "transaction_amount_distribution",
		Help:      "Distribution of transaction amounts",
		Buckets:   []float64{10, 50, 100, 500, 1000, 5000, 10000},
	}, []string{"type"})

	// User metrics
	activeUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "banking",
		Name:      "active_users",
		Help:      "Number of currently active users",
	})

	// Account metrics
	accountBalance = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "banking",
		Name:      "account_balance_distribution",
		Help:      "Distribution of account balances",
		Buckets:   []float64{100, 1000, 5000, 10000, 50000, 100000},
	}, []string{"currency"})

	// System metrics
	dbConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "banking",
		Name:      "db_connections_active",
		Help:      "Number of active database connections",
	})

	apiLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "banking",
		Name:      "api_request_duration_seconds",
		Help:      "API request latency distribution",
		Buckets:   prometheus.DefBuckets,
	}, []string{"endpoint", "method"})

	// Login metrics
	loginAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "banking",
		Name:      "login_attempts_total",
		Help:      "Total number of login attempts",
	}, []string{"status"})

	// Account metrics
	accountsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "banking",
		Name:      "accounts_total",
		Help:      "Total number of accounts by currency",
	}, []string{"currency"})

	dbStats = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "banking",
		Name:      "db_stats",
		Help:      "Database statistics",
	}, []string{"stat"})

	// SOA metrics
	soaGeneration = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "banking",
		Name:      "soa_generation_total",
		Help:      "Total number of statement of account generations",
	}, []string{"status"})

	soaGenerationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "banking",
		Name:      "soa_generation_duration_seconds",
		Help:      "Time taken to generate statement of account",
		Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
	})
)

func init() {
	// Register any non-promauto metrics
	prometheus.MustRegister()
}

// RecordRequest records API request metrics
func RecordRequest(path, method string, status int, duration float64) {
	// Convert status code to a readable string
	statusStr := fmt.Sprintf("%d", status) // e.g., "200", "400", "500"
	
	// Record the request
	requestsTotal.WithLabelValues(path, method, statusStr).Inc()
	apiLatency.WithLabelValues(path, method).Observe(duration)
}

// RecordTransaction records transaction metrics
func RecordTransaction(transactionType string, amount float64) {
	transactionCounter.WithLabelValues(transactionType).Inc()
	transactionAmount.WithLabelValues(transactionType).Observe(amount)
}

// IncrementActiveUsers increments the active users count
func IncrementActiveUsers() {
	activeUsers.Inc()
}

// DecrementActiveUsers decrements the active users count
func DecrementActiveUsers() {
	activeUsers.Dec()
}

// UpdateActiveUsers sets the active users count to a specific value
// This should only be used for initialization or recovery
func UpdateActiveUsers(count int) {
	activeUsers.Set(float64(count))
}

// RecordAccountBalance records account balance metrics
func RecordAccountBalance(balance float64, currency string) {
	accountBalance.WithLabelValues(currency).Observe(balance)
}

// UpdateDBConnections updates the database connections gauge
func UpdateDBConnections(connections int) {
	dbConnections.Set(float64(connections))
}

// RecordLoginAttempt records login attempts
func RecordLoginAttempt(success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	loginAttempts.WithLabelValues(status).Inc()
}

// RecordNewAccount records new account creation
func RecordNewAccount(currency string) {
	accountsTotal.WithLabelValues(currency).Inc()
}

// RecordDBStats records various database statistics
func RecordDBStats(stats map[string]float64) {
	for stat, value := range stats {
		dbStats.WithLabelValues(stat).Set(value)
	}
}

// RecordSOAGeneration records a SOA generation attempt
func RecordSOAGeneration(success bool, duration float64) {
	status := "success"
	if !success {
		status = "failure"
	}
	soaGeneration.WithLabelValues(status).Inc()
	soaGenerationDuration.Observe(duration)
}