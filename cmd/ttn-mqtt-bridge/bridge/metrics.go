package bridge

import "github.com/prometheus/client_golang/prometheus"

var gatewaysConnected = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "bridge",
		Name:      "gateway_connections_total",
		Help:      "Total number of connections from each gateway.",
	},
	[]string{"gateway_id"},
)

var gatewaysDisconnected = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "bridge",
		Name:      "gateway_disconnections_total",
		Help:      "Total number of disconnections from each gateway.",
	},
	[]string{"gateway_id"},
)

func init() {
	prometheus.MustRegister(gatewaysConnected, gatewaysDisconnected)
}
