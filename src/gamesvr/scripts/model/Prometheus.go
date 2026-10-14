package model

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	PrometheusRegistry *prometheus.Registry

	HttpHandler http.Handler

	OnlineNum = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ac_server_online_num",
		Help: "online player num",
	})

	LoginFailTimes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ac_server_login_failed_times",
		Help: "player login failed times",
	})

	GoRoutinueNum = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ac_server_gamesvr_goroutine_num",
		Help: "gamesvr goroutine num",
	})
)

func InitPrometheus() {
	PrometheusRegistry = prometheus.NewRegistry()
	PrometheusRegistry.Register(OnlineNum)
	PrometheusRegistry.Register(LoginFailTimes)
	PrometheusRegistry.Register(GoRoutinueNum)

	HttpHandler = promhttp.HandlerForTransactional(prometheus.ToTransactionalGatherer(PrometheusRegistry), promhttp.HandlerOpts{})
	//PrometheusRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}
