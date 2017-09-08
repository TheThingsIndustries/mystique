// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import "io"

// PubcompPacket is the PUBCOMP packet
type PubcompPacket struct {
	PacketIdentifier uint16
}

// PacketType returns the MQTT packet type of this packet
func (PubcompPacket) PacketType() byte { return PUBCOMP }

func (p *PubcompPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p PubcompPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p PubcompPacket) MarshalBinary() (data []byte, err error) {
	return encodeUint16(p.PacketIdentifier), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *PubcompPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return io.EOF
	}
	p.PacketIdentifier = decodeUint16(data)
	return nil
}

// Validate the packet contents
func (p PubcompPacket) Validate() error { return validatePacketIdentifier(p.PacketIdentifier) }
