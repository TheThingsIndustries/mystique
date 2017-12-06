// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

// EventMetadata for server events
type EventMetadata struct {
	Topic string `json:"topic,omitempty"`
}

func (s *serverSession) PublishEvent(name string, e EventMetadata) {
	go func() {
		pkt := &packet.PublishPacket{
			Received:  time.Now(),
			TopicName: fmt.Sprintf("$SYS/session/%s/%s", s.ID(), name),
		}
		pkt.Message, _ = json.Marshal(e)
		s.filteredDeliveryMu.Lock()
		defer s.filteredDeliveryMu.Unlock()
		if s.filteredDelivery != nil {
			s.filteredDelivery <- pkt
		}
	}()
}
