// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import "github.com/prometheus/client_golang/prometheus"

var stores []Store

var sessionsGauge = prometheus.NewGaugeFunc(
	prometheus.GaugeOpts{
		Namespace: "mystique",
		Name:      "sessions",
		Help:      "Number of sessions.",
	},
	func() (total float64) {
		for _, store := range stores {
			total += float64(len(store.All()))
		}
		return
	},
)

func init() {
	prometheus.MustRegister(sessionsGauge)
}
