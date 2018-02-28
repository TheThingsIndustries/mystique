// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package pending

import (
	"testing"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPendingPackets(t *testing.T) {
	a := assertions.New(t)
	p := new(List)

	a.So(p.Get(), should.BeEmpty)

	pkt := &packet.PublishPacket{TopicName: "foo"}
	a.So(p.Add(0, pkt), should.BeTrue)
	a.So(p.Add(0, pkt), should.BeFalse)
	a.So(p.Add(1, new(packet.PublishPacket)), should.BeTrue)
	a.So(p.Add(2, new(packet.PublishPacket)), should.BeTrue)
	a.So(p.Add(3, new(packet.PublishPacket)), should.BeTrue)

	a.So(p.Get(), should.HaveLength, 4)
	a.So(p.Get(), should.Contain, pkt)

	a.So(p.Remove(1), should.BeTrue)
	a.So(p.Get(), should.HaveLength, 3)
	a.So(p.Get(), should.Contain, pkt)

	a.So(p.Remove(0), should.BeTrue)
	a.So(p.Get(), should.HaveLength, 2)
	a.So(p.Get(), should.NotContain, pkt)

	a.So(p.Remove(0), should.BeFalse)
	a.So(p.Get(), should.HaveLength, 2)

	p.Clear()
	a.So(p.Get(), should.BeEmpty)
}
