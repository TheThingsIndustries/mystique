// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import (
	"bytes"
	"io"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// UnsubscribePacket is the UNSUBSCRIBE packet
type UnsubscribePacket struct {
	PacketIdentifier uint16
	Topics           []string
}

// PacketType returns the MQTT packet type of this packet
func (UnsubscribePacket) PacketType() byte { return UNSUBSCRIBE }

func (p *UnsubscribePacket) setFlags(f flags) error { return validateFlags(p.flags(), f) }

func (p UnsubscribePacket) flags() flags { return flags{false, true, false, false} }

// MarshalBinary implements encoding.BinaryMarshaler
func (p UnsubscribePacket) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	WriteUint16(buf, p.PacketIdentifier)
	for _, topic := range p.Topics {
		WriteString(buf, topic)
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *UnsubscribePacket) UnmarshalBinary(data []byte) (err error) {
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
		p.Topics = append(p.Topics, topic)
	}
	return nil
}

// Response to the packet
func (p UnsubscribePacket) Response() *UnsubackPacket {
	return &UnsubackPacket{PacketIdentifier: p.PacketIdentifier}
}

// Validate the packet contents
func (p UnsubscribePacket) Validate() (err error) {
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
