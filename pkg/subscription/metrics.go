// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package subscription

import "github.com/prometheus/client_golang/prometheus"

var subscriptionsGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "mystique",
		Name:      "subscriptions",
		Help:      "Number of subscriptions.",
	},
)

func init() {
	prometheus.MustRegister(subscriptionsGauge)
}
