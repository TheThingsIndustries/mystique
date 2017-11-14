// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import (
	"bytes"
	"io"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// SubscribePacket is the SUBSCRIBE packet
type SubscribePacket struct {
	PacketIdentifier uint16
	Topics           []string
	QoSs             []byte
}

// PacketType returns the MQTT packet type of this packet
func (SubscribePacket) PacketType() byte { return SUBSCRIBE }

func (p *SubscribePacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p SubscribePacket) flags() flags { return flags{false, true, false, false} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p SubscribePacket) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	WriteUint16(buf, p.PacketIdentifier)
	for i, topic := range p.Topics {
		WriteString(buf, topic)
		WriteByte(buf, p.QoSs[i]&0x3)
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *SubscribePacket) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	p.PacketIdentifier, err = ReadUint16(buf)
	if err != nil {
		return
	}
	for buf.Len() > 0 {
		topic, err := ReadString(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		qos, err := ReadByte(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		p.Topics = append(p.Topics, topic)
		switch qos {
		case 0, 1, 2:
			p.QoSs = append(p.QoSs, qos)
		default:
			return ErrProtocolViolation
		}
	}
	return nil
}

// Response to the packet
func (p SubscribePacket) Response() *SubackPacket {
	returnCodes := make([]byte, len(p.QoSs))
	copy(returnCodes, p.QoSs)
	return &SubackPacket{PacketIdentifier: p.PacketIdentifier, ReturnCodes: returnCodes}
}

// SubscribeRejected response code
var SubscribeRejected byte = 0x80

// Validate the packet contents
func (p SubscribePacket) Validate() (err error) {
	if err = validatePacketIdentifier(p.PacketIdentifier); err != nil {
		return err
	}
	for _, t := range p.Topics {
		if err = topic.ValidateFilter(t); err != nil {
			return err
		}
	}
	return nil
}
