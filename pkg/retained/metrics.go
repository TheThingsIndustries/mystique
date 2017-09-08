// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package retained

import "github.com/prometheus/client_golang/prometheus"

var retainedGauge = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "mystique",
		Name:      "retained_messages",
		Help:      "Number of retained messages.",
	},
)

func init() {
	prometheus.MustRegister(retainedGauge)
}
