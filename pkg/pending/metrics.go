// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package pending

import "github.com/prometheus/client_golang/prometheus"

var pendingMessagesGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "mystique",
		Name:      "pending_messages",
		Help:      "Number of pending messages.",
	},
)

func init() {
	prometheus.MustRegister(pendingMessagesGauge)
}
