package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// BusinessMetrics 定义业务监控指标
type BusinessMetrics struct {
	UserRegisteredTotal    prometheus.Counter
	DepositAmountTotal     *prometheus.CounterVec
	WithdrawAmountTotal    *prometheus.CounterVec
	SweeperJobDuration     *prometheus.HistogramVec
	AddressPoolRemaining   *prometheus.GaugeVec
	WithdrawalSuccessTotal *prometheus.CounterVec
}

// Global Metrics Instance
var Business *BusinessMetrics

// InitBusinessMetrics 初始化业务指标
func InitBusinessMetrics() {
	Business = &BusinessMetrics{
		UserRegisteredTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "wallet_user_registered_total",
			Help: "The total number of registered users",
		}),
		DepositAmountTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "wallet_deposit_amount_total",
			Help: "The total amount of deposits",
		}, []string{"currency"}),
		WithdrawAmountTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "wallet_withdraw_amount_total",
			Help: "The total amount of withdraws",
		}, []string{"currency"}),
		SweeperJobDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "wallet_sweeper_job_duration_seconds",
			Help:    "Duration of sweeper jobs",
			Buckets: prometheus.DefBuckets,
		}, []string{"chain"}),
		AddressPoolRemaining: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "wallet_address_pool_remaining",
			Help: "Remaining addresses in the pool",
		}, []string{"currency"}),
		WithdrawalSuccessTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "wallet_withdraw_success_total",
			Help: "Total number of successful withdrawals",
		}, []string{"chain"}),
	}
}
