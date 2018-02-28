// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import "io"

// PubackPacket is the PUBACK packet
type PubackPacket struct {
	PacketIdentifier uint16
}

// PacketType returns the MQTT packet type of this packet
func (PubackPacket) PacketType() byte { return PUBACK }

func (p *PubackPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p PubackPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p PubackPacket) MarshalBinary() (data []byte, err error) {
	return encodeUint16(p.PacketIdentifier), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *PubackPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return io.EOF
	}
	p.PacketIdentifier = decodeUint16(data)
	return nil
}

// Validate the packet contents
func (p PubackPacket) Validate() error { return validatePacketIdentifier(p.PacketIdentifier) }
