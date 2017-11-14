// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package server implements a simple MQTT server.
package server

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/retained"
	"github.com/TheThingsIndustries/mystique/pkg/session"
)

// Option for the server
type Option func(s *server)

// WithAuth returns an option that sets the authentication
func WithAuth(auth auth.Interface) Option {
	return func(s *server) { s.auth = auth }
}

// WithSessionStore returns an option that sets store for sessions
func WithSessionStore(sess session.Store) Option {
	return func(s *server) { s.sessions = sess }
}

// WithRetainedMessagesStore returns an option that sets the store for retained messages
func WithRetainedMessagesStore(store retained.Store) Option {
	return func(s *server) { s.retainedMessages = store }
}

// Server interface
type Server interface {
	Handle(conn net.Conn)
	Sessions() session.Store
}

// New returns a new MQTT server
func New(ctx context.Context, option ...Option) Server {
	s := &server{ctx: ctx, logger: log.FromContext(ctx)}
	for _, opt := range option {
		opt(s)
	}
	if s.sessions == nil {
		s.sessions = session.SimpleStore(s.ctx)
	}
	if s.retainedMessages == nil {
		s.retainedMessages = retained.SimpleStore(s.ctx)
	}
	return s
}

type server struct {
	ctx              context.Context
	logger           log.Interface
	auth             auth.Interface // may be nil
	sessions         session.Store
	retainedMessages retained.Store
}

// Handle a connection
func (s *server) Handle(conn net.Conn) {
	logger := s.logger.WithField("remote_addr", conn.RemoteAddr().String())
	logger.Debug("Open connection")
	conns.Inc()
	defer func() {
		logger.Debug("Close connection")
		conns.Dec()
		conn.Close()
	}()

	session, err := s.HandleConnect(conn)
	if err != nil {
		return
	}

	sessionID := session.ID()
	logger = logger.WithField("client_id", sessionID)
	if session.Username() != "" {
		logger = logger.WithField("username", session.Username())
	}

	defer func() {
		if session.IsGarbage() {
			logger.Info("End session")
			s.sessions.Delete(sessionID) // session.ID() is already empty at this point
		} else {
			logger.Info("Detach session")
		}
	}()

	// Make sure the session is closed when the client disconnects
	var kicked bool
	defer func() {
		if !kicked {
			session.Close()
		}
	}()

	// deliverLoop
	go func() {
		for msg := range session.DeliveryChan() {
			logger.WithFields(log.F{"topic": msg.TopicName, "size": len(msg.Message), "qos": msg.QoS}).Info("Publish message")
			s.sessions.Publish(msg)
		}
	}()

	// readLoop
	control := make(chan packet.ControlPacket)
	readErr := make(chan error, 1)
	go func() (err error) {
		defer func() {
			if err != nil {
				readErr <- err
			}
			close(readErr)
			close(control)
		}()
		for {
			pkt, err := conn.Receive()
			if err != nil {
				if err == io.EOF || kicked {
					return nil
				}
				logger.WithError(err).Warn("Error receiving packet")
				return err
			}
			if err := pkt.Validate(); err != nil {
				logger.WithError(err).Warn("Invalid packet")
				return err
			}
			var response packet.ControlPacket
			switch pkt := pkt.(type) {
			case *packet.PublishPacket:
				pkt.Received = time.Now().UTC()
				logger.Debug("Read PUBLISH")
				receivedCounter.WithLabelValues("publish").Inc()
				s.retainedMessages.Retain(pkt)
				response, err = session.HandlePublish(pkt)
			case *packet.PubackPacket:
				logger.Debug("Read PUBACK")
				receivedCounter.WithLabelValues("puback").Inc()
				err = session.HandlePuback(pkt)
			case *packet.PubrecPacket:
				logger.Debug("Read PUBREC")
				receivedCounter.WithLabelValues("pubrec").Inc()
				response, err = session.HandlePubrec(pkt)
			case *packet.PubrelPacket:
				logger.Debug("Read PUBREL")
				receivedCounter.WithLabelValues("pubrel").Inc()
				response, err = session.HandlePubrel(pkt)
			case *packet.PubcompPacket:
				logger.Debug("Read PUBCOMP")
				receivedCounter.WithLabelValues("pubcomp").Inc()
				err = session.HandlePubcomp(pkt)
			case *packet.SubscribePacket:
				logger.Debug("Read SUBSCRIBE")
				receivedCounter.WithLabelValues("subscribe").Inc()
				response, err = session.HandleSubscribe(pkt)
				for _, pkt := range s.retainedMessages.Get(pkt.Topics...) {
					session.Publish(pkt)
				}
			case *packet.UnsubscribePacket:
				logger.Debug("Read UNSUBSCRIBE")
				receivedCounter.WithLabelValues("unsubscribe").Inc()
				response, err = session.HandleUnsubscribe(pkt)
			case *packet.PingreqPacket:
				logger.Debug("Read PINGREQ")
				receivedCounter.WithLabelValues("pingreq").Inc()
				response = pkt.Response()
			case *packet.DisconnectPacket:
				logger.Debug("Read DISCONNECT")
				receivedCounter.WithLabelValues("disconnect").Inc()
				session.HandleDisconnect()
			default:
				logger.WithField("packet_type", fmt.Sprintf("%T", pkt)).Debug("Read unknown packet")
				receivedCounter.WithLabelValues("unknown").Inc()
				conn.Close()
			}
			if err != nil {
				logger.WithError(err).Warn("Could not handle packet")
				return err
			}
			if response != nil {
				switch response.(type) {
				case *packet.ConnackPacket:
					logger.Debug("Write CONNACK")
					sentCounter.WithLabelValues("connack").Inc()
				case *packet.PubackPacket:
					logger.Debug("Write PUBACK")
					sentCounter.WithLabelValues("puback").Inc()
				case *packet.PubrecPacket:
					logger.Debug("Write PUBREC")
					sentCounter.WithLabelValues("pubrec").Inc()
				case *packet.PubrelPacket:
					logger.Debug("Write PUBREL")
					sentCounter.WithLabelValues("pubrel").Inc()
				case *packet.PubcompPacket:
					logger.Debug("Write PUBCOMP")
					sentCounter.WithLabelValues("pubcomp").Inc()
				case *packet.SubackPacket:
					logger.Debug("Write SUBACK")
					sentCounter.WithLabelValues("suback").Inc()
				case *packet.UnsubackPacket:
					logger.Debug("Write UNSUBACK")
					sentCounter.WithLabelValues("unsuback").Inc()
				case *packet.PingrespPacket:
					logger.Debug("Write PINGRESP")
					sentCounter.WithLabelValues("pingresp").Inc()
				case *packet.DisconnectPacket:
					logger.Debug("Write DISCONNECT")
					sentCounter.WithLabelValues("disconnect").Inc()
				default:
					panic("trying to send unknown packet type")
				}
				control <- response
			}
		}
	}()

	// mainLoop
	publish := session.PublishChan()
	for {
		select {
		case readErr, ok := <-readErr:
			if ok {
				err = readErr
			}
			return
		case pkt, ok := <-control:
			if !ok {
				control = nil
				continue
			}
			err = conn.Send(pkt)
		case pkt, ok := <-publish:
			if !ok {
				kicked = true
				return
			}
			logger := logger
			if !pkt.Retain {
				latency := time.Since(pkt.Received)
				logger = logger.WithField("latency", latency)
				publishLatency.Observe(latency.Seconds())
			}
			logger.Debug("Write PUBLISH")
			sentCounter.WithLabelValues("publish").Inc()
			err = conn.Send(pkt)
		}
		if err != nil {
			return
		}
	}
}

func (s *server) Sessions() session.Store {
	return s.sessions
}

var boot = time.Now()

func (s *server) HandleConnect(conn net.Conn) (session session.ServerSession, err error) {
	logger := s.logger.WithField("remote_addr", conn.RemoteAddr().String())

	// Receive the connect packet or exit
	pkt, err := conn.Receive()
	if err != nil {
		return
	}
	connect, ok := pkt.(*packet.ConnectPacket)
	if !ok {
		logger.Warn("First packet is not CONNECT")
		return
	}

	defer func() {
		if err != nil {
			if code, ok := err.(packet.ConnectReturnCode); ok {
				response := connect.Response()
				response.ReturnCode = code
				conn.Send(response)
			}
		}
	}()

	// Validate the contents of the connect packet or exit with a return code
	if err = connect.Validate(); err != nil {
		logger.WithError(err).Warn("Invalid Connect packet")
		connectCounter.WithLabelValues("invalid").Inc()
		return
	}

	logger = logger.WithField("client_id", connect.ClientID)
	if connect.Username != "" {
		logger = logger.WithField("username", connect.Username)
	}

	authInfo := &auth.Info{
		RemoteAddr: conn.RemoteAddr().String(),
		Transport:  conn.Transport(),
		ClientID:   connect.ClientID,
		Username:   connect.Username,
		Password:   connect.Password,
	}

	if s.auth != nil {
		// Authenticate the connection or exit with a return code
		if err = s.auth.Connect(authInfo); err != nil {
			logger.WithError(err).Warn("Authentication failed")
			connectCounter.WithLabelValues("auth_refused").Inc()
			return
		}
	}

	if connect.KeepAlive > 0 {
		conn.SetReadTimeout(time.Duration(connect.KeepAlive) * 1500 * time.Millisecond)
	}

	if connect.ClientID == "" {
		connect.ClientID = fmt.Sprintf("%s-%d", conn.RemoteAddr().String(), time.Since(boot))
	}

	// Find a matching session or create a new one
	session = s.sessions.GetOrCreate(connect.ClientID)

	// Attach the connection or exit with a return code
	response, err := session.HandleConnect(conn, authInfo, connect)
	if err != nil {
		logger.WithError(err).Warn("Session failed to attach")
		connectCounter.WithLabelValues("session_refused").Inc()
		return
	}

	if response.SessionPresent {
		logger.Info("Attach session")
	} else {
		logger.Info("Start session")
	}
	connectCounter.WithLabelValues("accepted").Inc()

	// Send the connack
	conn.Send(response)

	// (Re)send pending messages
	for _, pkt := range session.Pending() {
		if pkt, ok := pkt.(*packet.PublishPacket); ok {
			pkt.Duplicate = true
		}
		conn.Send(pkt)
	}

	// Send retained messages
	for _, pkt := range s.retainedMessages.Get(session.SubscriptionTopics()...) {
		session.Publish(pkt)
	}

	return session, nil
}
