// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

// EventMetadata for server events
type EventMetadata struct {
	RemoteAddr string `json:"remote_addr,omitempty"`
	ClientID   string `json:"client_id,omitempty"`
	Username   string `json:"username,omitempty"`
}

func (s *server) PublishEvent(name string, e EventMetadata) {
	pkt := &packet.PublishPacket{
		Received:  time.Now(),
		TopicName: fmt.Sprintf("$SYS/server/events/%s", name),
	}
	pkt.Message, _ = json.Marshal(e)
	s.sessions.Publish(pkt)
}
