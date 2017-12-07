// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package server

import (
	"encoding/json"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// EventMetadata for server events
type EventMetadata struct {
	RemoteAddr string `json:"remote_addr,omitempty"`
	ClientID   string `json:"client_id,omitempty"`
	Username   string `json:"username,omitempty"`
}

var eventTopic = []string{"$SYS", "server", "events"}

func (s *server) PublishEvent(name string, e EventMetadata) {
	go func() {
		pkt := &packet.PublishPacket{
			Received:   time.Now(),
			TopicParts: append(eventTopic, name),
		}
		pkt.TopicName = topic.Join(pkt.TopicParts)
		pkt.Message, _ = json.Marshal(e)
		s.sessions.Publish(pkt)
	}()
}
