// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"errors"
	"sync/atomic"

	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

func (s *session) Publish(pkt *packet.PublishPacket) {
	if s.subscriptions.Count() == 0 {
		return
	}
	if !s.auth.CanRead(pkt.TopicParts...) {
		return
	}
	qos, ok := s.subscriptions.Match(pkt.TopicParts...)
	if !ok {
		return
	}
	logger := log.FromContext(s.ctx).WithFields(log.F{"topic": pkt.TopicName, "size": len(pkt.Message), "qos": pkt.QoS})
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
	if pub.QoS > 0 {
		publishIdentifier := atomic.AddUint64(&s.publishIdentifier, 1)
		pub.PacketIdentifier = uint16(publishIdentifier)
		s.pendingOut.Add(pub.PacketIdentifier, pub)
		if s.pendingOut.Len() > PublishBufferSize*2 {
			s.pendingOut.Clear()
			logger.WithField("error", "Too many pending messages").Warn("Cleared pendingOut")
		}
	}
	select {
	case s.publish <- pub:
		atomic.AddUint64(&s.published, 1)
		logger.Debug("Publish message")
	default:
		logger.WithError(errors.New("connection too slow")).Warn("Drop message")
	}
}

func (s *session) Deliver(pkt *packet.PublishPacket) {
	if s.auth.CanWrite(pkt.TopicParts...) {
		log.FromContext(s.ctx).WithFields(log.F{"topic": pkt.TopicName, "size": len(pkt.Message), "qos": pkt.QoS}).Debug("Deliver message")
		atomic.AddUint64(&s.delivered, 1)
		s.deliver(pkt)
	}
}

func (s *session) HandlePublish(pkt *packet.PublishPacket) (response packet.ControlPacket, err error) {
	response = pkt.Response()
	if pkt.QoS == 2 {
		if !s.pendingIn.Add(pkt.PacketIdentifier, pkt) { // already seen this message
			return
		}
		if s.pendingIn.Len() > PublishBufferSize*2 {
			s.pendingIn.Clear()
			log.FromContext(s.ctx).WithField("error", "Too many pending messages").Warn("Cleared pendingIn")
		}
	}
	s.Deliver(pkt)
	return
}

func (s *session) HandlePuback(pkt *packet.PubackPacket) (err error) {
	s.pendingOut.Remove(pkt.PacketIdentifier)
	return
}

func (s *session) HandlePubrec(pkt *packet.PubrecPacket) (response *packet.PubrelPacket, err error) {
	response = pkt.Response()
	s.pendingOut.Add(pkt.PacketIdentifier, response)
	return
}

func (s *session) HandlePubrel(pkt *packet.PubrelPacket) (response *packet.PubcompPacket, err error) {
	response = pkt.Response()
	s.pendingIn.Remove(pkt.PacketIdentifier)
	return
}

func (s *session) HandlePubcomp(pkt *packet.PubcompPacket) (err error) {
	s.pendingOut.Remove(pkt.PacketIdentifier)
	return
}

func (s *session) Pending() []packet.ControlPacket {
	return s.pendingOut.Get()
}
