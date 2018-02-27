// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package session implements MQTT sessions.
package session

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/pending"
	"github.com/TheThingsIndustries/mystique/pkg/subscription"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// PublishBufferSize sets the size of publish channel buffers
var PublishBufferSize = 64

// Session interface
type Session interface {
	Context() context.Context

	AuthInfo() auth.Info

	PublishChan() <-chan *packet.PublishPacket

	Stats() Stats

	// Read and handle the connect packet
	ReadConnect() error

	// Read and handle the next control packet, optionally returning a response
	ReadPacket() (packet.ControlPacket, error)

	// Handle a Disconnect packet
	// unsets the will
	HandleDisconnect()

	// Send an outgoing Publish message if the session is subscribed to the topic
	// if subscription with QoS 0: sends the message
	// if subscription with QoS 1: sends the message, stores the message until Puback
	// if subscription with QoS 2: sends the message, stores the message until Pubrec
	// if authentication is enabled, the server checks if the client is allowed to receive on the topic
	Publish(pkt *packet.PublishPacket)

	// Handle an incoming Publish packet
	// if QoS 0: delivers the packet and returns nil
	// if QoS 1: delivers the packet and returns a *PubackPacket
	// if QoS 2: delivers the packet, stores Pubrec until Pubrel, returns *PubrecPacket
	// if authentication is enabled, the server checks if the client is allowed to publish on the topic
	HandlePublish(pkt *packet.PublishPacket) (packet.ControlPacket, error)

	// Handle a Puback packet
	// clears pkt that was waiting for Puback
	HandlePuback(pkt *packet.PubackPacket) error

	// Handle a Pubrec packet
	// clears pkt that was waiting for Pubrec, stores Pubrel until Pubcomp, returns *PubrelPacket
	HandlePubrec(pkt *packet.PubrecPacket) (*packet.PubrelPacket, error)

	// Handle a Pubrel packet
	// clears Pubrec that was waiting for Pubrel, returns *PubcompPacket
	HandlePubrel(pkt *packet.PubrelPacket) (*packet.PubcompPacket, error)

	// Handle a Pubcomp packet
	// clears Pubrel that was waiting for Pubcomp
	HandlePubcomp(pkt *packet.PubcompPacket) error

	// Pending messages that should be retransmitted on a reconnect
	Pending() []packet.ControlPacket

	// Handle a Subscribe packet
	// adds subscriptions, returns *SubackPacket
	// if authentication is enabled, the server checks if the client is allowed to subscribe to the topic
	HandleSubscribe(pkt *packet.SubscribePacket) (*packet.SubackPacket, error)

	// Handle an Unsubscribe packet
	// removes subscriptions, returns *UnsubackPacket
	HandleUnsubscribe(pkt *packet.UnsubscribePacket) (*packet.UnsubackPacket, error)

	// Subscriptions of the session
	Subscriptions() map[string]byte

	// Close the session
	// closes the connection
	// delivers the will (if set) and then unsets it
	// clears the session state
	Close()
}

func New(ctx context.Context, conn net.Conn, deliver func(*packet.PublishPacket)) Session {
	return &session{
		ctx:     log.NewContext(ctx, log.FromContext(ctx)),
		conn:    conn,
		publish: make(chan *packet.PublishPacket, PublishBufferSize),
		deliver: deliver,
	}
}

type session struct {
	// BEGIN sync/atomic aligned
	publishIdentifier uint64
	published         uint64
	delivered         uint64
	// END sync/atomig aligned

	ctx     context.Context
	conn    net.Conn
	publish chan *packet.PublishPacket
	deliver func(pkt *packet.PublishPacket)

	auth *auth.Info

	// will of the session
	// can be set on (re)connect
	// is delivered when conn breaks
	// is cleared on HandleDisconnect
	will *packet.PublishPacket

	// pendingOut contains
	// - Publish packets that have not been sent
	// - Publish packets that have not been acknowledged with a Puback or Pubrec
	// - Pubrel packets that have not been acknowledged with a Pubcomp
	pendingOut pending.List

	// pendingIn contains
	// - Pubrec messages that have not been acknowledged with a Pubrel
	pendingIn pending.List

	// subcriptions of the session
	subscriptions subscription.List
}

func (s *session) Context() context.Context { return s.ctx }

func (s *session) AuthInfo() auth.Info {
	auth := *s.auth
	return auth
}

func (s *session) Stats() Stats {
	return Stats{
		Published: atomic.LoadUint64(&s.published),
		Delivered: atomic.LoadUint64(&s.delivered),
	}
}

func (s *session) PublishChan() <-chan *packet.PublishPacket {
	return s.publish
}

func (s *session) Close() {
	if s.will != nil {
		s.Deliver(s.will)
		s.will = nil
	}
	s.pendingOut.Clear()
	s.pendingIn.Clear()
	s.subscriptions.Clear()
}

func (s *session) ReadPacket() (response packet.ControlPacket, err error) {
	logger := log.FromContext(s.ctx)
	pkt, err := s.conn.Receive()
	if err != nil {
		if err != io.EOF {
			logger.WithError(err).Warn("Error receiving packet")
		}
		return nil, err
	}
	if err := pkt.Validate(); err != nil {
		logger.WithError(err).Warn("Received invalid packet")
		return nil, err
	}
	logger.Debugf("Read %s packet", packet.Name[pkt.PacketType()])
	switch pkt := pkt.(type) {
	case *packet.PublishPacket:
		pkt.Received = time.Now().UTC()
		if pkt.TopicName != "" {
			pkt.TopicParts = topic.Split(pkt.TopicName)
		}
		response, err = s.HandlePublish(pkt)
	case *packet.PubackPacket:
		err = s.HandlePuback(pkt)
	case *packet.PubrecPacket:
		response, err = s.HandlePubrec(pkt)
	case *packet.PubrelPacket:
		response, err = s.HandlePubrel(pkt)
	case *packet.PubcompPacket:
		err = s.HandlePubcomp(pkt)
	case *packet.SubscribePacket:
		response, err = s.HandleSubscribe(pkt)
	case *packet.UnsubscribePacket:
		response, err = s.HandleUnsubscribe(pkt)
	case *packet.PingreqPacket:
		response = pkt.Response()
	case *packet.DisconnectPacket:
		s.HandleDisconnect()
	default:
		err = errors.New("unknown packet type")
	}
	return
}
