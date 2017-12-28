// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import (
	"bytes"
	"errors"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// QoS Levels
const (
	AtMostOnce  = 0
	AtLeastOnce = 1
	ExactlyOnce = 2
)

// PublishPacket is the PUBLISH packet
type PublishPacket struct {
	Received         time.Time `json:"received"`
	Retain           bool      `json:"retained"`
	QoS              byte      `json:"qos"`
	Duplicate        bool      `json:"-"`
	PacketIdentifier uint16    `json:"-"`
	TopicName        string    `json:"topic"`
	TopicParts       []string  `json:"-"`
	Message          []byte    `json:"message"`
}

// PacketType returns the MQTT packet type of this packet
func (PublishPacket) PacketType() byte { return PUBLISH }

// ErrInvalidQoS is returned if the QoS of a publish packet is invalid
var ErrInvalidQoS = errors.New("Invalid QoS")

func (p *PublishPacket) setFlags(f flags) error {
	p.Retain = f[0]
	if f[1] {
		p.QoS |= (1 << 0)
	}
	if f[2] {
		p.QoS |= (1 << 1)
	}
	if p.QoS == 3 {
		return ErrInvalidQoS
	}
	p.Duplicate = f[3]
	return nil
}

func (p PublishPacket) flags() (f flags) {
	f[0] = p.Retain
	f[1] = p.QoS&1 == 1
	f[2] = p.QoS>>1&1 == 1
	f[3] = p.Duplicate
	return
}

// MarshalBinary implements encoding.BinaryMarshaler
func (p PublishPacket) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	WriteString(buf, p.TopicName)
	if p.QoS > 0 {
		WriteUint16(buf, p.PacketIdentifier)
	}
	buf.Write(p.Message)
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *PublishPacket) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	p.TopicName, err = ReadString(buf)
	if err != nil {
		return
	}
	if p.QoS > 0 {
		p.PacketIdentifier, err = ReadUint16(buf)
		if err != nil {
			return
		}
	}
	if buf.Len() > 0 {
		p.Message = buf.Bytes()
	}
	return
}

// Response to the packet
func (p PublishPacket) Response() ControlPacket {
	switch p.QoS {
	case 1:
		return &PubackPacket{PacketIdentifier: p.PacketIdentifier}
	case 2:
		return &PubrecPacket{PacketIdentifier: p.PacketIdentifier}
	}
	return nil
}

// Validate the packet contents
func (p PublishPacket) Validate() error {
	if p.QoS == 0 && p.Duplicate {
		return errors.New("DUP can not be 1 for QoS 0 messages")
	}
	return topic.ValidateTopic(p.TopicName)
}
