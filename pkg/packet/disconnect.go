// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

// DisconnectPacket is the DISCONNECT packet
type DisconnectPacket struct {
}

// PacketType returns the MQTT packet type of this packet
func (DisconnectPacket) PacketType() byte { return DISCONNECT }

func (p *DisconnectPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p DisconnectPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p DisconnectPacket) MarshalBinary() (data []byte, err error) { return nil, nil }

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *DisconnectPacket) UnmarshalBinary(data []byte) error { return nil }

// Validate the packet contents (noop)
func (p DisconnectPacket) Validate() error { return nil }
