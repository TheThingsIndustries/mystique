// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import "bytes"

// ConnectPacket is the CONNECT packet
type ConnectPacket struct {
	ProtocolName  string
	ProtocolLevel byte
	CleanStart    bool
	Will          bool
	WillQoS       byte
	WillRetain    bool
	KeepAlive     uint16
	ClientID      string
	WillTopic     string
	WillMessage   []byte
	Username      string
	Password      []byte
}

// PacketType returns the MQTT packet type of this packet
func (ConnectPacket) PacketType() byte { return CONNECT }

func (p *ConnectPacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p ConnectPacket) flags() flags { return flags{} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p ConnectPacket) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	WriteString(buf, p.ProtocolName)
	WriteByte(buf, p.ProtocolLevel)
	var flags byte
	flags |= bit(p.CleanStart) << 1
	flags |= bit(p.Will) << 2
	flags |= (p.WillQoS & 3) << 3
	flags |= bit(p.WillRetain) << 5
	flags |= bit(len(p.Password) > 0) << 6
	flags |= bit(len(p.Username) > 0) << 7
	WriteByte(buf, flags)
	WriteUint16(buf, p.KeepAlive)
	WriteString(buf, p.ClientID)
	if p.Will {
		WriteString(buf, p.WillTopic)
		WriteBytes(buf, p.WillMessage)
	}
	if len(p.Username) > 0 {
		WriteString(buf, p.Username)
	}
	if len(p.Password) > 0 {
		WriteBytes(buf, p.Password)
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *ConnectPacket) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	p.ProtocolName, err = ReadString(buf)
	if err != nil {
		return
	}
	p.ProtocolLevel, err = ReadByte(buf)
	if err != nil {
		return
	}
	var b byte
	b, err = ReadByte(buf)
	if err != nil {
		return
	}
	flags := newFlags(b)
	if flags[0] {
		return ErrProtocolViolation
	}
	p.CleanStart = flags[1]
	p.Will = flags[2]
	p.WillQoS = 3 & (b >> 3)
	if p.WillQoS == 3 {
		return ErrInvalidQoS
	}
	p.WillRetain = flags[5]
	passwordFlag := flags[6]
	usernameFlag := flags[7]
	p.KeepAlive, err = ReadUint16(buf)
	if err != nil {
		return
	}
	p.ClientID, err = ReadString(buf)
	if err != nil {
		return
	}
	if p.Will {
		p.WillTopic, err = ReadString(buf)
		if err != nil {
			return
		}
		p.WillMessage, err = ReadBytes(buf)
		if err != nil {
			return
		}
	}
	if usernameFlag {
		p.Username, err = ReadString(buf)
		if err != nil {
			return
		}
	}
	if passwordFlag {
		p.Password, err = ReadBytes(buf)
		if err != nil {
			return
		}
	}
	return
}

// Response to the packet
func (p ConnectPacket) Response() *ConnackPacket {
	return &ConnackPacket{}
}

// Validate the packet contents
func (p ConnectPacket) Validate() error {
	// TODO
	return nil
}
