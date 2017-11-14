// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package server

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type mockConn struct {
	received []packet.ControlPacket
	toSend   []packet.ControlPacket

	readTimeout time.Duration
	wg          sync.WaitGroup
	closed      bool
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}
func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1337}
}
func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1337}
}
func (m *mockConn) Transport() string {
	return "mock"
}
func (m *mockConn) Send(pkt packet.ControlPacket) error {
	if m.closed {
		return errors.New("conn closed")
	}
	if pkt, ok := pkt.(*packet.PublishPacket); ok {
		pkt.Received = time.Time{}
	}
	m.received = append(m.received, pkt)
	return nil
}
func (m *mockConn) Receive() (packet.ControlPacket, error) {
	if m.closed {
		return nil, errors.New("conn closed")
	}
	if len(m.toSend) > 0 {
		time.Sleep(50 * time.Millisecond)
		pkt := m.toSend[0]
		m.toSend = m.toSend[1:]
		return pkt, nil
	}
	m.wg.Wait()
	return nil, io.EOF
}
func (m *mockConn) SetReadTimeout(d time.Duration) {
	m.readTimeout = d
}

func newServer() *server {
	return New(context.Background(), WithAuth(nil), WithRetainedMessagesStore(nil), WithSessionStore(nil)).(*server)
}

func TestServer(t *testing.T) {
	a := assertions.New(t)

	s := newServer()
	a.So(s.auth, should.BeNil)                // Auth can be empty
	a.So(s.sessions, should.NotBeNil)         // Should use default
	a.So(s.retainedMessages, should.NotBeNil) // Should use default
}

func TestHandle(t *testing.T) {
	a := assertions.New(t)

	{
		s := newServer()
		conn := &mockConn{}
		s.Handle(conn)
		a.So(conn.closed, should.BeTrue)
		a.So(conn.received, should.BeEmpty)
	}

	{
		s := newServer()
		conn := &mockConn{}
		conn.toSend = append(conn.toSend,
			&packet.ConnectPacket{CleanStart: true},
		)
		s.Handle(conn)
		a.So(conn.closed, should.BeTrue)
		if a.So(conn.received, should.HaveLength, 1) {
			a.So(conn.received[0], should.HaveSameTypeAs, &packet.ConnackPacket{})
		}
	}

	{
		s := newServer()
		conn := &mockConn{}
		conn.toSend = append(conn.toSend,
			&packet.ConnectPacket{ProtocolName: "MQTT", ProtocolLevel: 4, CleanStart: true, KeepAlive: 1},
			&packet.PingreqPacket{},
			&packet.DisconnectPacket{},
		)
		s.Handle(conn)
		a.So(conn.closed, should.BeTrue)
		a.So(conn.readTimeout, should.Equal, 1500*time.Millisecond)
		a.So(conn.received, should.Resemble, []packet.ControlPacket{
			&packet.ConnackPacket{},
			&packet.PingrespPacket{},
		})
	}

	{
		s := newServer()

		conn0 := &mockConn{}
		conn0.wg.Add(1)
		conn0.toSend = append(conn0.toSend,
			&packet.ConnectPacket{ProtocolName: "MQTT", ProtocolLevel: 4, CleanStart: true},
			&packet.SubscribePacket{PacketIdentifier: 1, Topics: []string{"#"}, QoSs: []byte{0}},
		)
		go s.Handle(conn0)

		conn1 := &mockConn{}
		conn1.wg.Add(1)
		conn1.toSend = append(conn1.toSend,
			&packet.ConnectPacket{ProtocolName: "MQTT", ProtocolLevel: 4, CleanStart: true},
			&packet.SubscribePacket{PacketIdentifier: 1, Topics: []string{"#"}, QoSs: []byte{1}},
			// Not sending PUBACK
		)
		go s.Handle(conn1)

		conn2 := &mockConn{}
		conn2.wg.Add(1)
		conn2.toSend = append(conn2.toSend,
			&packet.ConnectPacket{ProtocolName: "MQTT", ProtocolLevel: 4, CleanStart: true},
			&packet.SubscribePacket{PacketIdentifier: 1, Topics: []string{"#"}, QoSs: []byte{2}},
			// Not sending PUBACK / PUBREC
		)
		go s.Handle(conn2)

		time.Sleep(100 * time.Millisecond)

		conn := &mockConn{}
		conn.toSend = append(conn.toSend,
			&packet.ConnectPacket{ProtocolName: "MQTT", ProtocolLevel: 4, CleanStart: true},
			&packet.SubscribePacket{PacketIdentifier: 41, Topics: []string{"foo"}, QoSs: []byte{1}},
			&packet.UnsubscribePacket{PacketIdentifier: 42, Topics: []string{"foo"}},
			&packet.PublishPacket{TopicName: "bar", QoS: 0},
			&packet.PublishPacket{PacketIdentifier: 43, TopicName: "bar", QoS: 1},
			&packet.PublishPacket{PacketIdentifier: 44, TopicName: "bar", QoS: 2},
			&packet.PubrelPacket{PacketIdentifier: 44},
			&packet.DisconnectPacket{},
		)
		s.Handle(conn)
		a.So(conn.closed, should.BeTrue)
		a.So(conn.received, should.Resemble, []packet.ControlPacket{
			&packet.ConnackPacket{},
			&packet.SubackPacket{PacketIdentifier: 41, ReturnCodes: []byte{1}},
			&packet.UnsubackPacket{PacketIdentifier: 42},
			&packet.PubackPacket{PacketIdentifier: 43},
			&packet.PubrecPacket{PacketIdentifier: 44},
			&packet.PubcompPacket{PacketIdentifier: 44},
		})

		conn0.wg.Done()
		conn1.wg.Done()
		conn2.wg.Done()

		a.So(conn0.received, should.Resemble, []packet.ControlPacket{
			&packet.ConnackPacket{},
			&packet.SubackPacket{PacketIdentifier: 1, ReturnCodes: []byte{0}},
			&packet.PublishPacket{TopicName: "bar", QoS: 0},
			&packet.PublishPacket{TopicName: "bar", QoS: 0},
			&packet.PublishPacket{TopicName: "bar", QoS: 0},
		})

		a.So(conn1.received, should.Resemble, []packet.ControlPacket{
			&packet.ConnackPacket{},
			&packet.SubackPacket{PacketIdentifier: 1, ReturnCodes: []byte{1}},
			&packet.PublishPacket{TopicName: "bar", QoS: 0},
			&packet.PublishPacket{PacketIdentifier: 1, TopicName: "bar", QoS: 1},
			&packet.PublishPacket{PacketIdentifier: 2, TopicName: "bar", QoS: 1},
		})

		a.So(conn2.received, should.Resemble, []packet.ControlPacket{
			&packet.ConnackPacket{},
			&packet.SubackPacket{PacketIdentifier: 1, ReturnCodes: []byte{2}},
			&packet.PublishPacket{TopicName: "bar", QoS: 0},
			&packet.PublishPacket{PacketIdentifier: 1, TopicName: "bar", QoS: 1},
			&packet.PublishPacket{PacketIdentifier: 2, TopicName: "bar", QoS: 2},
		})

	}
}
