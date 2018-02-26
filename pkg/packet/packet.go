// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package packet defines the MQTT packet types
package packet

import (
	"encoding"
	"errors"
	"io"
)

// PacketType contains the MQTT packet type
type PacketType byte

// Packet types
const (
	_           PacketType = 0  // Reserved
	CONNECT                = 1  // Client request to connect to Server
	CONNACK                = 2  // Connect acknowledgment
	PUBLISH                = 3  // Publish message
	PUBACK                 = 4  // Publish acknowledgment
	PUBREC                 = 5  // Publish received (assured delivery part 1)
	PUBREL                 = 6  // Publish release (assured delivery part 2)
	PUBCOMP                = 7  // Publish complete (assured delivery part 3)
	SUBSCRIBE              = 8  // Subscribe request
	SUBACK                 = 9  // Subscribe acknowledgment
	UNSUBSCRIBE            = 10 // Unsubscribe request
	UNSUBACK               = 11 // Unsubscribe acknowledgment
	PINGREQ                = 12 // PING request
	PINGRESP               = 13 // PING response
	DISCONNECT             = 14 // Client is disconnecting
	AUTH                   = 15 // Authentication exchange
)

var Name = map[byte]string{
	CONNECT:     "connect",
	CONNACK:     "connack",
	PUBLISH:     "publish",
	PUBACK:      "puback",
	PUBREC:      "pubrec",
	PUBREL:      "pubrel",
	PUBCOMP:     "pubcomp",
	SUBSCRIBE:   "subscribe",
	SUBACK:      "suback",
	UNSUBSCRIBE: "unsubscribe",
	UNSUBACK:    "unsuback",
	PINGREQ:     "pingreq",
	PINGRESP:    "pingresp",
	DISCONNECT:  "disconnect",
	AUTH:        "auth",
}

// ErrProtocolViolation is returned when a message violates the protocol specification
var ErrProtocolViolation = errors.New("Protocol Violation")

// ControlPacket represents an MQTT Control Packet
type ControlPacket interface {
	encoding.BinaryMarshaler   // without fixed header
	encoding.BinaryUnmarshaler // without fixed header
	PacketType() byte
	setFlags(f flags) error
	flags() flags
	Validate() error
}

// Write a control packet to the writer
func Write(w io.Writer, p ControlPacket) (err error) {
	_, err = w.Write([]byte{p.PacketType()<<4 | p.flags().Byte()})
	if err != nil {
		return err
	}
	var payload []byte
	payload, err = p.MarshalBinary()
	if err != nil {
		return
	}
	err = WriteRemainingLength(w, len(payload))
	if err != nil {
		return
	}
	_, err = w.Write(payload)
	return
}

// ErrInvalidPacketType is returned when the control packet type is invalid
var ErrInvalidPacketType = errors.New("Invalid packet type")

// Read a control packet from the Reader
func Read(r io.Reader) (p ControlPacket, err error) {
	b := make([]byte, 1)
	_, err = r.Read(b)
	if err != nil {
		return
	}
	var (
		messageType = PacketType(b[0] >> 4)
		flags       = newFlags(b[0])
		length      int
		payload     []byte
	)
	length, err = ReadRemainingLength(r)
	if err != nil {
		return
	}
	if length > 0 {
		payload = make([]byte, length)
		_, err = r.Read(payload)
		if err != nil {
			return
		}
	}
	switch messageType {
	case CONNECT:
		p = new(ConnectPacket)
	case CONNACK:
		p = new(ConnackPacket)
	case PUBLISH:
		p = new(PublishPacket)
	case PUBACK:
		p = new(PubackPacket)
	case PUBREC:
		p = new(PubrecPacket)
	case PUBREL:
		p = new(PubrelPacket)
	case PUBCOMP:
		p = new(PubcompPacket)
	case SUBSCRIBE:
		p = new(SubscribePacket)
	case SUBACK:
		p = new(SubackPacket)
	case UNSUBSCRIBE:
		p = new(UnsubscribePacket)
	case UNSUBACK:
		p = new(UnsubackPacket)
	case PINGREQ:
		p = new(PingreqPacket)
	case PINGRESP:
		p = new(PingrespPacket)
	case DISCONNECT:
		p = new(DisconnectPacket)
	default:
		return nil, ErrInvalidPacketType
	}
	p.setFlags(flags)
	err = p.UnmarshalBinary(payload)
	return
}

// ErrInvalidPacketIdentifier is returned if an expected packet identifier was empty
var ErrInvalidPacketIdentifier = errors.New("Invalid packet identifier")

func validatePacketIdentifier(identifier uint16) error {
	if identifier == 0 {
		return ErrInvalidPacketIdentifier
	}
	return nil
}
