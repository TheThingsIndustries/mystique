// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

// PingrespPacket is the PINGRESP packet
type PingrespPacket struct {
}

// PacketType returns the MQTT packet type of this packet
func (PingrespPacket) PacketType() byte { return PINGRESP }

func (p *PingrespPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p PingrespPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p PingrespPacket) MarshalBinary() (data []byte, err error) { return nil, nil }

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *PingrespPacket) UnmarshalBinary(data []byte) error { return nil }

// Validate the packet contents (noop)
func (p PingrespPacket) Validate() error { return nil }
