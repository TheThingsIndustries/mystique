// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"encoding/json"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// EventMetadata for server events
type EventMetadata struct {
	Topic string `json:"topic,omitempty"`
}

var eventTopic = []string{"$SYS", "session"}

func (s *serverSession) PublishEvent(name string, e EventMetadata) {
	go func() {
		pkt := &packet.PublishPacket{
			Received:   time.Now(),
			TopicParts: append(eventTopic, s.ID(), name),
		}
		pkt.TopicName = topic.Join(pkt.TopicParts)
		pkt.Message, _ = json.Marshal(e)
		s.filteredDeliveryMu.Lock()
		defer s.filteredDeliveryMu.Unlock()
		if s.filteredDelivery != nil {
			s.filteredDelivery <- pkt
		}
	}()
}
