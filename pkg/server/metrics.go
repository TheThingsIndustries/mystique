// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package server

import "github.com/prometheus/client_golang/prometheus"

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
	prometheus.MustRegister(publishLatency)
	prometheus.MustRegister(conns)
}
