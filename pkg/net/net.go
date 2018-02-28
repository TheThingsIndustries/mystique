// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package net implements the MQTT network layer on top of the standard library "net" package.
package net

import (
	"context"
	"net"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

// Conn wraps net.Conn with mqtt-specific functions.
type Conn interface {
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	NetConn() net.Conn
	Transport() string
	Send(pkt packet.ControlPacket) error
	Receive() (packet.ControlPacket, error)
	SetReadTimeout(d time.Duration)
}

type conn struct {
	transport string
	timeout   time.Duration
	net.Conn
}

func (c *conn) NetConn() net.Conn {
	return c.Conn
}

func (c *conn) Transport() string {
	return c.transport
}

func (c *conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	sentBytes.Add(float64(n))
	return
}

func (c *conn) Send(pkt packet.ControlPacket) error {
	registerSend(pkt)
	return packet.Write(c, pkt)
}

func (c *conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	receivedBytes.Add(float64(n))
	return
}

func (c *conn) Receive() (packet.ControlPacket, error) {
	pkt, err := packet.Read(c)
	if err != nil {
		return nil, err
	}
	registerReceive(pkt)
	c.updateTimeout()
	return pkt, nil
}

func (c *conn) SetReadTimeout(d time.Duration) {
	c.timeout = d
	c.updateTimeout()
}

func (c *conn) updateTimeout() {
	var deadline time.Time
	if c.timeout > 0 {
		deadline = time.Now().Add(c.timeout)
	}
	c.SetReadDeadline(deadline)
}

// Dial connects to the address on the named network; see net.Dial for details.
func Dial(network, address string) (Conn, error) {
	return DialContext(context.Background(), network, address)
}

// NewConn creates a Conn that wraps an inner Conn.
func NewConn(inner net.Conn, transport string) Conn {
	return &conn{Conn: inner, transport: transport}
}

// DialContext acts like Dial but takes a context.
func DialContext(ctx context.Context, network, address string) (Conn, error) {
	var d net.Dialer
	inner, err := d.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return NewConn(inner, "tcp"), nil
}

// Listener wraps net.Listener with MQTT-specific functions.
type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
}

type listener struct {
	net.Listener
	transport string
}

func (l *listener) Accept() (Conn, error) {
	inner, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewConn(inner, l.transport), nil
}

// Listen on the local network address.
func Listen(network, address string) (Listener, error) {
	inner, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewListener(inner, network), nil
}

// NewListener creates a Listener that accepts connections from an inner Listener.
func NewListener(inner net.Listener, transport string) Listener {
	return &listener{Listener: inner, transport: transport}
}
