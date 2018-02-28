// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package pending implements storage for pending messages.
package pending

import (
	"sync"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

type pendingPacket struct {
	id  uint16
	pkt packet.ControlPacket
}

// List of pending messages
type List struct {
	mu       sync.Mutex
	messages []pendingPacket
}

// Add a pending packet to the end of the list
func (p *List) Add(id uint16, pkt packet.ControlPacket) (added bool) {
	p.mu.Lock()
	pending := pendingPacket{id: id, pkt: pkt}
	for i, pkt := range p.messages {
		if pkt.id == id {
			p.messages[i] = pending
			p.mu.Unlock()
			return false
		}
	}
	p.messages = append(p.messages, pending)
	pendingMessagesGauge.Inc()
	p.mu.Unlock()
	return true
}

// Remove a pending packet, guessing it is in the beginning of the list
func (p *List) Remove(id uint16) (removed bool) {
	p.mu.Lock()
	for i, pkt := range p.messages {
		if pkt.id == id {
			p.messages = append(p.messages[:i], p.messages[i+1:]...)
			removed = true
			pendingMessagesGauge.Dec()
			break
		}
	}
	p.mu.Unlock()
	return
}

// Clear the list
func (p *List) Clear() {
	p.mu.Lock()
	pendingMessagesGauge.Sub(float64(len(p.messages)))
	p.messages = nil
	p.mu.Unlock()
}

// Len returns the length of the list.
func (p *List) Len() (l int) {
	p.mu.Lock()
	l = len(p.messages)
	p.mu.Unlock()
	return
}

// Get all pending packets
func (p *List) Get() []packet.ControlPacket {
	p.mu.Lock()
	packets := make([]packet.ControlPacket, len(p.messages))
	for i, pkt := range p.messages {
		packets[i] = pkt.pkt
	}
	p.mu.Unlock()
	return packets
}
