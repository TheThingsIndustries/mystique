// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package server

import "github.com/prometheus/client_golang/prometheus"

var receivedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "server",
		Name:      "messages_received_total",
		Help:      "Total number of messages received.",
	},
	[]string{"message_type"},
)

var sentCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "server",
		Name:      "messages_sent_total",
		Help:      "Total number of messages sent.",
	},
	[]string{"message_type"},
)

var connectCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "server",
		Name:      "connect_handled_total",
		Help:      "Total connect messages handled.",
	},
	[]string{"result"},
)

var publishLatency = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "mystique",
		Subsystem: "server",
		Name:      "publish_latency_seconds",
		Help:      "Histogram of publish latency (seconds).",
		Buckets:   []float64{0.00025, 0.0005, 0.001, 0.0025, .005, .01, .025, .05, .1, .25, .5, 1},
	},
)

var conns = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "mystique",
	Subsystem: "server",
	Name:      "connections",
	Help:      "Number of server connections.",
})

func init() {
	prometheus.MustRegister(receivedCounter)
	prometheus.MustRegister(sentCounter)
	prometheus.MustRegister(connectCounter)
	prometheus.MustRegister(publishLatency)
}
