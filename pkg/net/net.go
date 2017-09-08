// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

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
	Send(pkt packet.ControlPacket) error
	Receive() (packet.ControlPacket, error)
	SetReadTimeout(d time.Duration)
}

type conn struct {
	timeout time.Duration
	net.Conn
}

func (c *conn) Send(pkt packet.ControlPacket) error {
	return packet.Write(c.Conn, pkt)
}

func (c *conn) Receive() (packet.ControlPacket, error) {
	pkt, err := packet.Read(c.Conn)
	if err != nil {
		return nil, err
	}
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
func NewConn(inner net.Conn) Conn {
	return &conn{Conn: inner}
}

// DialContext acts like Dial but takes a context.
func DialContext(ctx context.Context, network, address string) (Conn, error) {
	var d net.Dialer
	inner, err := d.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return NewConn(inner), nil
}

// Listener wraps net.Listener with MQTT-specific functions.
type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
}

type listener struct {
	net.Listener
}

func (l *listener) Accept() (Conn, error) {
	inner, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewConn(inner), nil
}

// Listen on the local network address.
func Listen(network, address string) (Listener, error) {
	inner, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewListener(inner), nil
}

// NewListener creates a Listener that accepts connections from an inner Listener.
func NewListener(inner net.Listener) Listener {
	return &listener{inner}
}
