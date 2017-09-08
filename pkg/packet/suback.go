// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import "bytes"

// SubackPacket is the SUBACK packet
type SubackPacket struct {
	PacketIdentifier uint16
	ReturnCodes      []byte
}

// PacketType returns the MQTT packet type of this packet
func (SubackPacket) PacketType() byte { return SUBACK }

func (p *SubackPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p SubackPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p SubackPacket) MarshalBinary() (data []byte, err error) {
	return append(encodeUint16(p.PacketIdentifier), p.ReturnCodes...), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *SubackPacket) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	p.PacketIdentifier, err = ReadUint16(buf)
	if l := buf.Len(); l > 0 {
		p.ReturnCodes = make([]byte, buf.Len())
		for i, code := range buf.Bytes() {
			switch code {
			case 0, 1, 2, 0x80:
				p.ReturnCodes[i] = code
			default:
				return ErrProtocolViolation
			}
		}
	}
	return nil
}

// Validate the packet contents
func (p SubackPacket) Validate() error {
	// TODO
	return validatePacketIdentifier(p.PacketIdentifier)
}
