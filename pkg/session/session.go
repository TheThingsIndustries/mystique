// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package session implements MQTT sessions.
package session

import (
	"context"
	"sync"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/pending"
	"github.com/TheThingsIndustries/mystique/pkg/subscription"
)

// PublishBufferSize sets the size of publish channel buffers
var PublishBufferSize = 64

// Session interface shared in both client and server
type Session interface {
	Context() context.Context

	ID() string
	Username() string

	// Delivery channel of incoming publish messages
	DeliveryChan() <-chan *packet.PublishPacket

	// Publish channel of outgoing publish messages
	PublishChan() <-chan *packet.PublishPacket

	// Send an outgoing Publish message
	// if subscription with QoS 0: sends the message
	// if subscription with QoS 1: sends the message, stores the message until Puback
	// if subscription with QoS 2: sends the message, stores the message until Pubrec
	//
	// the server-side implementation of Publish adds the following:
	// - it matches the topic of the message to the session's subscriptions before publishing
	// - it checks if the client is allowed to receive on the topic
	Publish(pkt *packet.PublishPacket)

	// Handle an incoming Publish packet
	// if QoS 0: delivers the packet and returns nil
	// if QoS 1: delivers the packet and returns a *PubackPacket
	// if QoS 2: delivers the packet, stores Pubrec until Pubrel, returns *PubrecPacket
	//
	// the server-side implementation of Publish adds the following:
	// - it checks if the client is allowed to publish on the topic
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

	Subscriptions() map[string]byte

	SubscriptionTopics() []string

	// Close the session
	// closes the connection (if connected)
	// on the server: delivers the will (if set) and then unsets it
	// on the server: clears the session state if CleanStart
	Close()
}

func newSession(ctx context.Context) session {
	return session{
		ctx:        ctx,
		logger:     log.FromContext(ctx),
		publishOut: make(chan *packet.PublishPacket, PublishBufferSize),
		delivery:   make(chan *packet.PublishPacket),
	}
}

type session struct {
	mu sync.RWMutex

	ctx context.Context

	logger log.Interface

	authMu   sync.RWMutex
	authinfo *auth.Info

	// (indirectly) contains the session ID and other options
	connect *packet.ConnectPacket

	// current publish packetIdentifier number; only the 16lsb will actually be used
	publishIdentifier uint64

	// publish send queue - buffered
	publishOut chan *packet.PublishPacket

	// publish delivery channel - blocking
	delivery chan *packet.PublishPacket

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

	wg sync.WaitGroup
}

func (s *session) Context() context.Context { return s.ctx }

func (s *session) getAuthInfo() *auth.Info {
	s.authMu.RLock()
	authinfo := s.authinfo
	s.authMu.RUnlock()
	return authinfo
}

func (s *session) setAuthInfo(authinfo *auth.Info) {
	s.authMu.Lock()
	s.authinfo = authinfo
	s.authMu.Unlock()
}

func (s *session) ID() (id string) {
	if authinfo := s.getAuthInfo(); authinfo != nil {
		id = authinfo.ClientID
	}
	return
}

func (s *session) Username() (username string) {
	if authinfo := s.getAuthInfo(); authinfo != nil {
		username = authinfo.Username
	}
	return
}
