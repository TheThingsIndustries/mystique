// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import "io"

// UnsubackPacket is the UNSUBACK packet
type UnsubackPacket struct {
	PacketIdentifier uint16
}

// PacketType returns the MQTT packet type of this packet
func (UnsubackPacket) PacketType() byte { return UNSUBACK }

func (p *UnsubackPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p UnsubackPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p UnsubackPacket) MarshalBinary() (data []byte, err error) {
	return encodeUint16(p.PacketIdentifier), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *UnsubackPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return io.EOF
	}
	p.PacketIdentifier = decodeUint16(data)
	return nil
}

// Validate the packet contents
func (p UnsubackPacket) Validate() error { return validatePacketIdentifier(p.PacketIdentifier) }
