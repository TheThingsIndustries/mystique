// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package retained implements a store for retained messages.
package retained

import (
	"context"
	"sync"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// Store for retained messages
type Store interface {
	// Retain the message if the RETAIN bit is set
	Retain(*packet.PublishPacket)

	// Get all currently retained messages that match the filters
	Get(filter ...string) []*packet.PublishPacket

	// All retained messages
	All() []*packet.PublishPacket
}

// SimpleStore returns a simple store for retained messages
func SimpleStore(ctx context.Context) Store {
	return &retainedMessages{
		ctx:      ctx,
		messages: make(map[string]*packet.PublishPacket),
	}
}

type retainedMessages struct {
	mu       sync.RWMutex
	ctx      context.Context
	messages map[string]*packet.PublishPacket
}

func (r *retainedMessages) Retain(pkt *packet.PublishPacket) {
	if !pkt.Retain {
		return
	}
	pkt.Retain = false // Unset retain flag on original message
	r.mu.Lock()
	_, exists := r.messages[pkt.TopicName]
	if len(pkt.Message) > 0 {
		retained := *pkt
		retained.Retain = true // Set retain flag on message copy
		r.messages[pkt.TopicName] = &retained
		if !exists {
			retainedGauge.Inc()
		}
	} else if exists {
		delete(r.messages, pkt.TopicName)
		retainedGauge.Dec()
	}
	r.mu.Unlock()
}

func (r *retainedMessages) Get(filter ...string) (packets []*packet.PublishPacket) {
nextRetained:
	for _, pkt := range r.All() {
		for _, filter := range filter {
			if topic.Match(pkt.TopicName, filter) {
				packets = append(packets, pkt)
				continue nextRetained
			}
		}
	}
	return
}

func (r *retainedMessages) All() []*packet.PublishPacket {
	r.mu.RLock()
	defer r.mu.RUnlock()
	packets := make([]*packet.PublishPacket, 0, len(r.messages))
	for _, packet := range r.messages {
		packets = append(packets, packet)
	}
	return packets
}
