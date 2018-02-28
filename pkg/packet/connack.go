// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import "bytes"

// ConnackPacket is the CONNACK packet
type ConnackPacket struct {
	ReturnCode     ConnectReturnCode
	SessionPresent bool
}

// PacketType returns the MQTT packet type of this packet
func (ConnackPacket) PacketType() byte { return CONNACK }

func (p *ConnackPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p ConnackPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p ConnackPacket) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	var flags byte
	flags |= bit(p.SessionPresent)
	WriteByte(buf, flags)
	WriteByte(buf, byte(p.ReturnCode))
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *ConnackPacket) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	var flags flags
	flags, err = readFlags(buf)
	if err != nil {
		return
	}
	p.SessionPresent = flags[0]
	if flags[1] || flags[2] || flags[3] || flags[4] || flags[5] || flags[6] || flags[7] {
		return ErrProtocolViolation
	}
	var returnCode byte
	returnCode, err = ReadByte(buf)
	if err != nil {
		return
	}
	p.ReturnCode = ConnectReturnCode(returnCode)
	if !p.ReturnCode.valid() {
		return ErrProtocolViolation
	}
	return
}

// Validate the packet contents
func (p ConnackPacket) Validate() error {
	// TODO
	return nil
}
