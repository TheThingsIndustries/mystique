// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package net

import (
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/prometheus/client_golang/prometheus"
)

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

var receivedMessages = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "server",
		Name:      "messages_received_total",
		Help:      "Total number of messages received.",
	},
	[]string{"message_type"},
)

var sentMessages = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "mystique",
		Subsystem: "server",
		Name:      "messages_sent_total",
		Help:      "Total number of messages sent.",
	},
	[]string{"message_type"},
)

func init() {
	prometheus.MustRegister(receivedBytes)
	prometheus.MustRegister(sentBytes)
	prometheus.MustRegister(receivedMessages)
	prometheus.MustRegister(sentMessages)
}

func registerSend(pkt packet.ControlPacket) {
	packetType := packet.Name[pkt.PacketType()]
	if packetType == "" {
		packetType = "unknown"
	}
	sentMessages.WithLabelValues(packetType).Inc()
}

func registerReceive(pkt packet.ControlPacket) {
	packetType := packet.Name[pkt.PacketType()]
	if packetType == "" {
		packetType = "unknown"
	}
	receivedMessages.WithLabelValues(packetType).Inc()
}
