// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package packet

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var connectTestSubjects = []ControlPacket{
	&ConnectPacket{},
	&ConnectPacket{CleanStart: true},
	&ConnectPacket{Will: true, WillMessage: []byte{1, 2, 3, 4}, WillQoS: 1, WillRetain: true, WillTopic: "foo"},
	&ConnectPacket{Username: "username", Password: []byte("password")},
}

var connackTestSubjects = []ControlPacket{
	&ConnackPacket{SessionPresent: true},
	&ConnackPacket{ReturnCode: ConnectNotAuthorized},
}

var publishTestSubjects = []ControlPacket{
	&PublishPacket{QoS: 0},
	&PublishPacket{QoS: 1, PacketIdentifier: 1},
	&PublishPacket{QoS: 2, PacketIdentifier: 1},
}

var pubackTestSubjects = []ControlPacket{
	&PubackPacket{
		PacketIdentifier: 1,
	},
}

var pubrecTestSubjects = []ControlPacket{
	&PubrecPacket{
		PacketIdentifier: 1,
	},
}

var pubrelTestSubjects = []ControlPacket{
	&PubrelPacket{
		PacketIdentifier: 1,
	},
}

var pubcompTestSubjects = []ControlPacket{
	&PubcompPacket{
		PacketIdentifier: 1,
	},
}

var subscribeTestSubjects = []ControlPacket{
	&SubscribePacket{
		PacketIdentifier: 1,
		Topics:           []string{"foo"},
		QoSs:             []byte{0x01},
	},
}

var subackTestSubjects = []ControlPacket{
	&SubackPacket{
		PacketIdentifier: 1,
		ReturnCodes:      []byte{0x01},
	},
}

var unsubscribeTestSubjects = []ControlPacket{
	&UnsubscribePacket{
		PacketIdentifier: 1,
		Topics:           []string{"foo"},
	},
}

var unsubackTestSubjects = []ControlPacket{
	&UnsubackPacket{
		PacketIdentifier: 1,
	},
}

var pingreqTestSubjects = []ControlPacket{
	&PingreqPacket{},
}

var pingrespTestSubjects = []ControlPacket{
	&PingrespPacket{},
}

var disconnectTestSubjects = []ControlPacket{
	&DisconnectPacket{},
}

func TestMarshalUnmarshal(t *testing.T) {
	var subjects = [][]ControlPacket{
		connectTestSubjects,
		connackTestSubjects,
		publishTestSubjects,
		pubackTestSubjects,
		pubrecTestSubjects,
		pubrelTestSubjects,
		pubcompTestSubjects,
		subscribeTestSubjects,
		subackTestSubjects,
		unsubscribeTestSubjects,
		unsubackTestSubjects,
		pingreqTestSubjects,
		pingrespTestSubjects,
		disconnectTestSubjects,
	}

	buf := new(bytes.Buffer)

	a := assertions.New(t)
	for _, pkt := range subjects {
		if len(pkt) == 0 {
			continue
		}
		t.Run(fmt.Sprintf("%T", pkt[0]), func(t *testing.T) {
			for _, pkt := range pkt {
				buf.Reset()
				a.So(Write(buf, pkt), should.BeNil)
				read, err := Read(buf)
				a.So(err, should.BeNil)
				a.So(read, should.Resemble, pkt)
			}
		})
	}
}

func TestResponse(t *testing.T) {
	a := assertions.New(t)
	a.So((&ConnectPacket{}).Response(), should.HaveSameTypeAs, &ConnackPacket{})
	a.So((&PublishPacket{QoS: 0}).Response(), should.BeNil)
	a.So((&PublishPacket{QoS: 1}).Response(), should.HaveSameTypeAs, &PubackPacket{})
	a.So((&PublishPacket{QoS: 2}).Response(), should.HaveSameTypeAs, &PubrecPacket{})
	a.So((&PubrecPacket{}).Response(), should.HaveSameTypeAs, &PubrelPacket{})
	a.So((&PubrelPacket{}).Response(), should.HaveSameTypeAs, &PubcompPacket{})
	a.So((&SubscribePacket{}).Response(), should.HaveSameTypeAs, &SubackPacket{})
	a.So((&UnsubscribePacket{}).Response(), should.HaveSameTypeAs, &UnsubackPacket{})
	a.So((&PingreqPacket{}).Response(), should.HaveSameTypeAs, &PingrespPacket{})
}
