// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"testing"

	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPublishingQoS0(t *testing.T) {
	a := assertions.New(t)

	sessionA := &session{publishOut: make(chan *packet.PublishPacket, 16), delivery: make(chan *packet.PublishPacket, 1), logger: log.Noop}
	sessionB := &session{publishOut: make(chan *packet.PublishPacket, 16), delivery: make(chan *packet.PublishPacket, 1), logger: log.Noop}

	msg := &packet.PublishPacket{
		TopicName: "foo",
		Message:   []byte("bar"),
	}

	sessionA.Publish(msg)

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 0)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)

	res, err := sessionB.HandlePublish(<-sessionA.PublishChan())
	a.So(err, should.BeNil)
	a.So(res, should.BeNil)

	a.So(sessionB.DeliveryChan(), should.HaveLength, 1)
	a.So(<-sessionB.DeliveryChan(), should.Resemble, msg)

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 0)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)
}

func TestPublishingQoS1(t *testing.T) {
	a := assertions.New(t)

	sessionA := &session{publishOut: make(chan *packet.PublishPacket, 16), delivery: make(chan *packet.PublishPacket, 1), logger: log.Noop}
	sessionB := &session{publishOut: make(chan *packet.PublishPacket, 16), delivery: make(chan *packet.PublishPacket, 1), logger: log.Noop}

	msg := &packet.PublishPacket{
		TopicName: "foo",
		Message:   []byte("bar"),
		QoS:       1,
	}

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 0)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)

	sessionA.Publish(msg)

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 1)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)

	puback, err := sessionB.HandlePublish(<-sessionA.PublishChan())
	a.So(err, should.BeNil)
	a.So(puback, should.NotBeNil)
	a.So(puback, should.HaveSameTypeAs, new(packet.PubackPacket))

	a.So(sessionB.DeliveryChan(), should.HaveLength, 1)
	a.So(<-sessionB.DeliveryChan(), should.Resemble, msg)

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 1)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)

	err = sessionA.HandlePuback(puback.(*packet.PubackPacket))
	a.So(err, should.BeNil)

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 0)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)
}

func TestPublishingQoS2(t *testing.T) {
	a := assertions.New(t)

	sessionA := &session{publishOut: make(chan *packet.PublishPacket, 16), delivery: make(chan *packet.PublishPacket, 1), logger: log.Noop}
	sessionB := &session{publishOut: make(chan *packet.PublishPacket, 16), delivery: make(chan *packet.PublishPacket, 1), logger: log.Noop}

	msg := &packet.PublishPacket{
		TopicName: "foo",
		Message:   []byte("bar"),
		QoS:       2,
	}

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 0)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)

	sessionA.Publish(msg)

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 1)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)

	pub := <-sessionA.PublishChan()

	pubrec, err := sessionB.HandlePublish(pub)
	a.So(err, should.BeNil)
	a.So(pubrec, should.NotBeNil)
	a.So(pubrec, should.HaveSameTypeAs, new(packet.PubrecPacket))

	a.So(sessionB.DeliveryChan(), should.HaveLength, 1)
	a.So(<-sessionB.DeliveryChan(), should.Resemble, msg)

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 1)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 1)

	{ // Retx should not redeliver
		pubrec, err := sessionB.HandlePublish(pub)
		a.So(err, should.BeNil)
		a.So(pubrec, should.NotBeNil)
		a.So(pubrec, should.HaveSameTypeAs, new(packet.PubrecPacket))

		a.So(sessionB.DeliveryChan(), should.HaveLength, 0)

		a.So(sessionA.pendingOut.Get(), should.HaveLength, 1)
		a.So(sessionB.pendingIn.Get(), should.HaveLength, 1)
	}

	pubrel, err := sessionA.HandlePubrec(pubrec.(*packet.PubrecPacket))
	a.So(err, should.BeNil)
	a.So(pubrel, should.NotBeNil)
	a.So(pubrel, should.HaveSameTypeAs, new(packet.PubrelPacket))

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 1)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 1)

	pubcomp, err := sessionB.HandlePubrel(pubrel)
	a.So(err, should.BeNil)
	a.So(pubcomp, should.NotBeNil)
	a.So(pubcomp, should.HaveSameTypeAs, new(packet.PubcompPacket))

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 1)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)

	err = sessionA.HandlePubcomp(pubcomp)
	a.So(err, should.BeNil)

	a.So(sessionA.pendingOut.Get(), should.HaveLength, 0)
	a.So(sessionB.pendingIn.Get(), should.HaveLength, 0)
}
