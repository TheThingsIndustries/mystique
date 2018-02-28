// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

// PingreqPacket is the PINGREQ packet
type PingreqPacket struct {
}

// PacketType returns the MQTT packet type of this packet
func (PingreqPacket) PacketType() byte { return PINGREQ }

func (p *PingreqPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p PingreqPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p PingreqPacket) MarshalBinary() (data []byte, err error) { return nil, nil }

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *PingreqPacket) UnmarshalBinary(data []byte) error { return nil }

// Response to the packet
func (p PingreqPacket) Response() *PingrespPacket {
	return &PingrespPacket{}
}

// Validate the packet contents (noop)
func (p PingreqPacket) Validate() error { return nil }
