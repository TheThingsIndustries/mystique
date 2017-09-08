// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import "github.com/prometheus/client_golang/prometheus"

var sessionsGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "mystique",
		Name:      "sessions",
		Help:      "Number of sessions.",
	},
)

func init() {
	prometheus.MustRegister(sessionsGauge)
}
