// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import "io"

// PubrelPacket is the PUBREL packet
type PubrelPacket struct {
	PacketIdentifier uint16
}

// PacketType returns the MQTT packet type of this packet
func (PubrelPacket) PacketType() byte { return PUBREL }

func (p *PubrelPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p PubrelPacket) flags() flags { return flags{false, true, false, false} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p PubrelPacket) MarshalBinary() (data []byte, err error) {
	return encodeUint16(p.PacketIdentifier), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *PubrelPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return io.EOF
	}
	p.PacketIdentifier = decodeUint16(data)
	return nil
}

// Response to the packet
func (p PubrelPacket) Response() *PubcompPacket {
	return &PubcompPacket{PacketIdentifier: p.PacketIdentifier}
}

// Validate the packet contents
func (p PubrelPacket) Validate() error { return validatePacketIdentifier(p.PacketIdentifier) }
