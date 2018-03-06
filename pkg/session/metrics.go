// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import "github.com/prometheus/client_golang/prometheus"

var stores []*simpleStore

var sessionsGauge = prometheus.NewGaugeFunc(
	prometheus.GaugeOpts{
		Namespace: "mystique",
		Name:      "sessions",
		Help:      "Number of sessions.",
	},
	func() (total float64) {
		for _, store := range stores {
			total += float64(store.Count())
		}
		return
	},
)

var sessionDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "mystique",
		Subsystem: "sessions",
		Name:      "duration_seconds",
		Help:      "Duration of the session (measured at disconnect).",
		Buckets: []float64{
			1, 10, 30,
			60, 60 * 5, 60 * 10, 60 * 30,
			3600, 3600 * 6, 3600 * 12, 3600 * 24,
		},
	},
)

var sessionMessages = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "mystique",
		Subsystem: "sessions",
		Name:      "messages",
		Help:      "Number of messages sent during the session (measured at disconnect).",
		Buckets:   []float64{0, 1, 5, 10, 100, 500, 1000, 5000, 10000, 50000, 100000},
	},
	[]string{"direction"},
)

func init() {
	prometheus.MustRegister(sessionsGauge)
	prometheus.MustRegister(sessionDuration)
	prometheus.MustRegister(sessionMessages)
}
