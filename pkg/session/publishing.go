// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

func (s *session) Publish(pkt *packet.PublishPacket) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pkt.QoS > 0 {
		s.publishIdentifier++
		pkt.PacketIdentifier = uint16(s.publishIdentifier)
		s.pendingOut.Add(pkt.PacketIdentifier, pkt)
	}
	logger := s.logger.WithFields(log.Fields("topic", pkt.TopicName, "size", len(pkt.Message), "qos", pkt.QoS))
	select {
	case s.publishOut <- pkt:
		logger.Debug("Publish message")
	default:
		logger.Warn("Dropping message [buffer full]")
		// TODO: if session connected, this means the client can't keep up, what to do?
		// TODO: if the session is disconnected, we should probably just discard the session
	}
}

func (s *session) HandlePublish(pkt *packet.PublishPacket) (response packet.ControlPacket, err error) {
	response = pkt.Response()
	if pkt.QoS == 2 {
		if !s.pendingIn.Add(pkt.PacketIdentifier, pkt) { // already seen this message
			return
		}
	}
	s.delivery <- pkt
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

func (s *session) DeliveryChan() (delivery <-chan *packet.PublishPacket) {
	return s.delivery
}

func (s *session) PublishChan() (publishOut <-chan *packet.PublishPacket) {
	return s.publishOut
}

func (s *session) Pending() []packet.ControlPacket {
	return s.pendingOut.Get()
}
