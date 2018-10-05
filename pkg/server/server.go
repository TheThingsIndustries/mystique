// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package server implements a simple MQTT server.
package server

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/session"
)

// Option for the server
type Option func(s *server)

// WithAuth returns an option that sets the authentication
func WithAuth(iface auth.Interface) Option {
	return func(s *server) {
		s.ctx = auth.NewContextWithInterface(s.ctx, iface)
	}
}

// WithSessionStore returns an option that sets store for sessions
func WithSessionStore(sess session.Store) Option {
	return func(s *server) { s.sessions = sess }
}

// WithIPLimits returns an option that sets limits on connections per IP
func WithIPLimits(max int) Option {
	return func(s *server) { s.ipLimits = newLimits(max) }
}

// WithUserLimits returns an option that sets limits on connections per User
func WithUserLimits(max int) Option {
	return func(s *server) { s.userLimits = newLimits(max) }
}

func WithAuthRevalidation(ttl time.Duration) Option {
	return func(s *server) { s.revalidateAuth = ttl }
}

// Server interface
type Server interface {
	Sessions() session.Store
	Publish(pkt *packet.PublishPacket)
	Handle(conn mqttnet.Conn)
}

// New returns a new MQTT server
func New(ctx context.Context, option ...Option) Server {
	s := &server{ctx: ctx}
	for _, opt := range option {
		opt(s)
	}
	if s.sessions == nil {
		s.sessions = session.SimpleStore()
	}
	return s
}

type server struct {
	ctx            context.Context
	ipLimits       *limits
	userLimits     *limits
	sessions       session.Store
	revalidateAuth time.Duration
}

func (s *server) Sessions() session.Store {
	return s.sessions
}

func (s *server) Publish(pkt *packet.PublishPacket) {
	s.sessions.Publish(pkt)
}

func (s *server) Handle(conn mqttnet.Conn) {
	s.handle(conn)
}

func (s *server) handle(conn mqttnet.Conn) (err error) {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	remoteAddr := conn.RemoteAddr().String()

	logger := log.FromContext(s.ctx).WithField("remote_addr", remoteAddr)
	ctx = log.NewContext(ctx, logger)

	logger.Debug("Open connection")
	conns.Inc()
	defer func() {
		if err != nil && err != io.EOF {
			logger = logger.WithError(err)
		}
		logger.Debug("Close connection")
		conns.Dec()
		conn.Close()
	}()

	ip, _, _ := net.SplitHostPort(remoteAddr)
	if err = s.ipLimits.connect(ip); err != nil {
		return err
	}
	defer s.ipLimits.disconnect(ip)

	session := session.New(ctx, conn, s.Publish)

	if err = session.ReadConnect(); err != nil {
		return err
	}
	defer session.Close()

	if username := session.AuthInfo().Username; username != "" {
		if err = s.userLimits.connect(username); err != nil {
			return err
		}
		defer s.userLimits.disconnect(username)
	}

	s.sessions.Store(session)
	defer s.sessions.Delete(session)

	logger = log.FromContext(session.Context()) // update with session fields

	control := make(chan packet.ControlPacket)
	readErr := make(chan error, 1)
	go func() {
		for {
			response, err := session.ReadPacket()
			if err != nil {
				readErr <- err
				close(readErr)
				return
			}
			if response != nil {
				logger.Debugf("Write %s packet", packet.Name[response.PacketType()])
				control <- response
			}
		}
	}()

	authInfo := session.AuthInfo()
	var revalidateAuth <-chan time.Time
	var authRevalidator auth.Revalidator
	if authInterface := auth.InterfaceFromContext(s.ctx); authInterface != nil {
		if revalidator, ok := authInterface.(auth.Revalidator); ok {
			authRevalidator = revalidator

			t := time.NewTicker(s.revalidateAuth)
			defer t.Stop()
			revalidateAuth = t.C
		}
	}

	// mainLoop
	publish := session.PublishChan()
	for {
		select {
		case <-revalidateAuth:
			err = authRevalidator.Revalidate(session.Context(), &authInfo)
		case readErr, ok := <-readErr:
			if ok {
				err = readErr
			}
			return err
		case pkt, ok := <-control:
			if !ok {
				control = nil
				continue
			}
			err = conn.Send(pkt)
		case pkt, ok := <-publish:
			if !ok {
				return
			}
			logger := logger
			if !pkt.Retain && !pkt.Duplicate && !pkt.Received.IsZero() {
				latency := time.Since(pkt.Received)
				logger = logger.WithField("latency", latency)
				publishLatency.Observe(latency.Seconds())
			}
			logger.Debug("Write publish packet")
			err = conn.Send(pkt)
		}
		if err != nil {
			return err
		}
	}
}
