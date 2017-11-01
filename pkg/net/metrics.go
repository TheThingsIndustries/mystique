// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package net

import "github.com/prometheus/client_golang/prometheus"

var receivedBytes = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "server",
		Name:      "bytes_received_total",
		Help:      "Total number of bytes received.",
	},
)

var sentBytes = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "server",
		Name:      "bytes_sent_total",
		Help:      "Total number of bytes sent.",
	},
)

func init() {
	prometheus.MustRegister(receivedBytes)
	prometheus.MustRegister(sentBytes)
}
