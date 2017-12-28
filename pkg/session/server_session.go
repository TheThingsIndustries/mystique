// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

type serverSession struct {
	// BEGIN align on 32 bit machines
	publishCount  uint64
	deliveryCount uint64
	// END align on 32 bit machines

	session

	expires time.Time

	filteredDeliveryMu sync.Mutex
	filteredDelivery   chan *packet.PublishPacket
}

func (s *serverSession) IsGarbage() (isGarbage bool) {
	s.mu.RLock()
	if !s.expires.IsZero() && s.expires.Before(time.Now()) {
		isGarbage = true
	}
	s.mu.RUnlock()
	return
}

func (s *serverSession) HandleConnect(conn net.Conn, authInfo *auth.Info, pkt *packet.ConnectPacket) (*packet.ConnackPacket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := &packet.ConnackPacket{}
	if s.connect != nil { // existing session being taken over
		if s.connect.Username != pkt.Username { // check that username matches
			return nil, packet.ConnectIdentifierRejected
		}
		if s.publishOut != nil { // existing session still connected; close that
			s.logger.Debug("Kick old connection")
			s.close()
		}
		if pkt.CleanStart { // want a clean session; reset state
			s.logger.Debug("Clean old session")
			s.clear()
		} else {
			response.SessionPresent = true
		}
	}

	s.connect = pkt
	s.setAuthInfo(authInfo)
	s.expires = time.Time{}

	s.logger = s.logger.WithFields(log.F{"client_id": authInfo.ClientID, "remote_addr": conn.RemoteAddr().String()})
	if authInfo.Username != "" {
		s.logger = s.logger.WithField("username", authInfo.Username)
	}

	if pkt.Will {
		s.will = &packet.PublishPacket{
			Retain:    pkt.WillRetain,
			QoS:       pkt.WillQoS,
			TopicName: pkt.WillTopic,
			Message:   pkt.WillMessage,
		}
		s.will.TopicParts = topic.Split(s.will.TopicName)
		s.PublishEvent("session.set_will", EventMetadata{Topic: s.will.TopicName})
	}

	return response, nil
}

func (s *serverSession) RemoteAddr() string {
	if authinfo := s.getAuthInfo(); authinfo != nil {
		return authinfo.RemoteAddr
	}
	return ""
}

func (s *serverSession) Stats() Stats {
	return Stats{
		Published: atomic.LoadUint64(&s.publishCount),
		Delivered: atomic.LoadUint64(&s.deliveryCount),
	}
}

func (s *serverSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.close()
}

func (s *serverSession) close() {
	if s.connect == nil {
		return
	}

	s.logger.Debug("Close session")

	publishOut := s.publishOut
	s.publishOut = make(chan *packet.PublishPacket, PublishBufferSize)
	close(publishOut)

	s.deliverWill()

	delivery := s.delivery
	s.delivery = make(chan *packet.PublishPacket)
	close(delivery)

	s.wg.Wait() // Wait for the goroutine to finish

	s.expires = time.Now().Add(time.Hour)

	if s.connect.CleanStart {
		s.clear()
	}
}

func (s *serverSession) clear() {
	if s.connect == nil {
		return
	}

	s.logger.Debug("Clear session")

	s.subscriptions.Clear()
	s.pendingIn.Clear()
	s.pendingOut.Clear()

	s.setAuthInfo(nil)
	s.connect = nil
	s.will = nil
	s.publishIdentifier = 0
	s.logger = log.FromContext(s.ctx)

	s.publishCount = 0
	s.deliveryCount = 0

	s.expires = time.Now()
}

func (s *serverSession) deliverWill() {
	if s.will != nil {
		s.delivery <- s.will
		s.will = nil
	}
}

func (s *serverSession) HandleDisconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.will = nil
}

func (s *serverSession) HandleSubscribe(pkt *packet.SubscribePacket) (*packet.SubackPacket, error) {
	response := pkt.Response()
	for i, topic := range pkt.Topics {
		logger := s.logger
		acceptedTopic, qos, err := s.getAuthInfo().Subscribe(topic, pkt.QoSs[i])
		if err != nil {
			response.ReturnCodes[i] = packet.SubscribeRejected
			s.PublishEvent("session.subscribe_rejected", EventMetadata{Topic: topic})
			continue
		}
		if acceptedTopic != topic {
			logger = logger.WithField("topic_original", topic)
		}
		if s.subscriptions.Add(acceptedTopic, qos) {
			logger.WithFields(log.F{"topic": acceptedTopic, "qos": qos}).Debug("Subscribe")
			s.PublishEvent("session.subscribed", EventMetadata{Topic: acceptedTopic})
		}
		response.ReturnCodes[i] = qos
	}
	return response, nil
}

func (s *serverSession) HandleUnsubscribe(pkt *packet.UnsubscribePacket) (*packet.UnsubackPacket, error) {
	response := pkt.Response()
	for _, topic := range pkt.Topics {
		if s.subscriptions.Remove(topic) {
			s.logger.WithField("topic", topic).Debug("Unsubscribe")
			s.PublishEvent("session.unsubscribed", EventMetadata{Topic: topic})
		}
	}
	return response, nil
}

func (s *serverSession) Subscriptions() map[string]byte {
	return s.subscriptions.Subscriptions()
}

func (s *serverSession) SubscriptionTopics() []string {
	return s.subscriptions.SubscriptionTopics()
}

func (s *serverSession) Publish(pkt *packet.PublishPacket) {
	if s.subscriptions.Count() == 0 {
		return
	}
	canRead := s.getAuthInfo().CanRead(pkt.TopicParts...)
	if !canRead {
		return
	}
	qos, ok := s.subscriptions.Match(pkt.TopicParts...)
	if !ok {
		return
	}
	pub := &packet.PublishPacket{
		Received:   pkt.Received,
		Retain:     pkt.Retain,
		QoS:        qos,
		TopicName:  pkt.TopicName,
		TopicParts: pkt.TopicParts,
		Message:    pkt.Message,
	}
	if pub.QoS > pkt.QoS {
		pub.QoS = pkt.QoS
	}
	atomic.AddUint64(&s.publishCount, 1)
	s.session.Publish(pub)
}

func (s *serverSession) DeliveryChan() <-chan *packet.PublishPacket {
	s.filteredDeliveryMu.Lock()
	if s.filteredDelivery == nil {
		s.mu.Lock()
		delivery := s.delivery
		s.mu.Unlock()
		s.filteredDelivery = make(chan *packet.PublishPacket)
		s.wg.Add(1)
		go func() {
			for pkt := range delivery {
				if s.getAuthInfo().CanWrite(pkt.TopicParts...) {
					atomic.AddUint64(&s.deliveryCount, 1)
					s.filteredDelivery <- pkt
				} else {
					s.logger.WithField("topic", pkt.TopicName).Debug("Drop unauthorized publish")
					s.PublishEvent("session.publish.refused", EventMetadata{Topic: pkt.TopicName})
				}
			}
			s.filteredDeliveryMu.Lock()
			defer s.filteredDeliveryMu.Unlock()
			close(s.filteredDelivery)
			s.filteredDelivery = nil
			s.wg.Done()
		}()
	}
	s.filteredDeliveryMu.Unlock()
	return s.filteredDelivery
}
