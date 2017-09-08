// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import "io"

// PubrecPacket is the PUBREC packet
type PubrecPacket struct {
	PacketIdentifier uint16
}

// PacketType returns the MQTT packet type of this packet
func (PubrecPacket) PacketType() byte { return PUBREC }

func (p *PubrecPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p PubrecPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p PubrecPacket) MarshalBinary() (data []byte, err error) {
	return encodeUint16(p.PacketIdentifier), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *PubrecPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return io.EOF
	}
	p.PacketIdentifier = decodeUint16(data)
	return nil
}

// Response to the packet
func (p PubrecPacket) Response() *PubrelPacket {
	return &PubrelPacket{PacketIdentifier: p.PacketIdentifier}
}

// Validate the packet contents
func (p PubrecPacket) Validate() error { return validatePacketIdentifier(p.PacketIdentifier) }
