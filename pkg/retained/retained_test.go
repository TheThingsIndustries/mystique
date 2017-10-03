// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package retained

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions/should"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/smartystreets/assertions"
)

func TestSubscriptions(t *testing.T) {
	a := assertions.New(t)
	s := SimpleStore(context.Background())

	a.So(s.Get("#"), should.HaveLength, 0)

	s.Retain(&packet.PublishPacket{
		TopicName: "foo",
		Retain:    true,
		Message:   []byte("foo"),
	})

	s.Retain(&packet.PublishPacket{
		TopicName: "bar",
		Retain:    true,
		Message:   []byte("bar"),
	})

	a.So(s.Get("#"), should.HaveLength, 2)

	s.Retain(&packet.PublishPacket{
		TopicName: "bar",
		Retain:    true,
	})

	a.So(s.Get("#"), should.HaveLength, 1)
}
