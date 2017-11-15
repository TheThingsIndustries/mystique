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

func (s *session) PublishEvent(name string, e EventMetadata) {
	pkt := &packet.PublishPacket{
		Received:  time.Now(),
		TopicName: fmt.Sprintf("$SYS/session/%s/%s", s.ID(), name),
	}
	pkt.Message, _ = json.Marshal(e)
	s.delivery <- pkt
}
