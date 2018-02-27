// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

func (s *session) HandleSubscribe(pkt *packet.SubscribePacket) (*packet.SubackPacket, error) {
	response := pkt.Response()
	logger := log.FromContext(s.ctx)
	for i, topic := range pkt.Topics {
		acceptedTopic, qos, err := s.auth.Subscribe(topic, pkt.QoSs[i])
		if err != nil {
			response.ReturnCodes[i] = packet.SubscribeRejected
			continue
		}
		logger := logger // shadow
		if acceptedTopic != topic {
			logger = logger.WithField("topic_original", topic)
		}
		if s.subscriptions.Add(acceptedTopic, qos) {
			logger.WithFields(log.F{"topic": acceptedTopic, "qos": qos}).Debug("Subscribe")
		}
		response.ReturnCodes[i] = qos
	}
	return response, nil
}

func (s *session) HandleUnsubscribe(pkt *packet.UnsubscribePacket) (*packet.UnsubackPacket, error) {
	response := pkt.Response()
	logger := log.FromContext(s.ctx)
	for _, topic := range pkt.Topics {
		acceptedTopic, _, err := s.auth.Subscribe(topic, 0)
		if err != nil {
			continue
		}
		logger := logger // shadow
		if acceptedTopic != topic {
			logger = logger.WithField("topic_original", topic)
		}
		if s.subscriptions.Remove(acceptedTopic) {
			logger.WithField("topic", acceptedTopic).Debug("Unsubscribe")
		}
	}
	return response, nil
}

func (s *session) Subscriptions() map[string]byte {
	return s.subscriptions.Subscriptions()
}
