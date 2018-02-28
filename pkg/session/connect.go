// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"crypto/tls"
	"errors"
	"fmt"
	gnet "net"
	"strings"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"golang.org/x/net/websocket"
)

var (
	boot            = time.Now()
	replaceClientID = strings.NewReplacer("/", ".")
)

func (s *session) ReadConnect() error {
	logger := log.FromContext(s.ctx)

	pkt, err := s.conn.Receive()
	if err != nil {
		return err
	}
	connectPacket, ok := pkt.(*packet.ConnectPacket)
	if !ok {
		return errors.New("First packet was not a CONNECT")
	}
	connackPacket := connectPacket.Response()

	err = connectPacket.Validate()
	if err != nil {
		logger.WithError(err).Warn("Invalid CONNECT")
		if code, ok := err.(packet.ConnectReturnCode); ok {
			connackPacket.ReturnCode = code
			if err := s.conn.Send(connackPacket); err != nil {
				logger.WithError(err).Warn("Could not send CONNACK")
				return err
			}
		}
		return err
	}

	if connectPacket.ClientID == "" {
		connectPacket.ClientID = fmt.Sprintf("%s-%d", s.conn.RemoteAddr().String(), time.Since(boot))
	} else {
		connectPacket.ClientID = replaceClientID.Replace(connectPacket.ClientID)
	}

	logger = logger.WithFields(log.F{
		"username":  connectPacket.Username,
		"client_id": connectPacket.ClientID,
	})

	s.auth = &auth.Info{
		RemoteAddr: s.conn.RemoteAddr().String(),
		Transport:  s.conn.Transport(),
		ClientID:   connectPacket.ClientID,
		Username:   connectPacket.Username,
		Password:   connectPacket.Password,
	}

	switch conn := s.conn.NetConn().(type) {
	case *gnet.TCPConn:
	case *tls.Conn:
		s.auth.ServerName = conn.ConnectionState().ServerName
	case *websocket.Conn:
		s.auth.ServerName = conn.Request().Host
	}

	if s.auth.ServerName != "" {
		logger = logger.WithField("server_name", s.auth.ServerName)
	}

	s.ctx = log.NewContext(s.ctx, logger)

	if authInterface := auth.InterfaceFromContext(s.ctx); authInterface != nil {
		if err = authInterface.Connect(s.auth); err != nil {
			if code, ok := err.(packet.ConnectReturnCode); ok {
				connackPacket.ReturnCode = code
				if err := s.conn.Send(connackPacket); err != nil {
					return err
				}
			}
			logger.WithError(err).Debug("Rejected authentication")
			return err
		}
	}

	if connectPacket.KeepAlive > 0 {
		s.conn.SetReadTimeout(time.Duration(connectPacket.KeepAlive) * 1500 * time.Millisecond)
	}

	if connectPacket.Will {
		topicParts := topic.Split(connectPacket.WillTopic)
		if s.auth.CanWrite(topicParts...) {
			s.will = &packet.PublishPacket{
				Retain:     connectPacket.WillRetain,
				QoS:        connectPacket.WillQoS,
				TopicName:  connectPacket.WillTopic,
				TopicParts: topicParts,
				Message:    connectPacket.WillMessage,
			}
		}
	}

	err = s.conn.Send(connackPacket)
	if err != nil {
		logger.WithError(err).Warn("Could not send CONNACK")
		return err
	}

	return nil
}

func (s *session) HandleDisconnect() {
	s.will = nil
}
