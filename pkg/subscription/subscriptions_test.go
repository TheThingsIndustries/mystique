// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package subscription

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSubscriptions(t *testing.T) {
	a := assertions.New(t)
	s := new(List)

	_, ok := s.Match("foo")
	a.So(ok, should.BeFalse)

	a.So(s.Add("foo", 1), should.BeTrue)

	a.So(s.Subscriptions(), should.ContainKey, "foo")
	a.So(s.SubscriptionTopics(), should.Contain, "foo")

	qos, ok := s.Match("foo")
	a.So(ok, should.BeTrue)
	a.So(qos, should.Equal, 1)

	a.So(s.Add("foo", 0), should.BeFalse)
	a.So(s.Add("+", 2), should.BeTrue)

	qos, ok = s.Match("foo")
	a.So(ok, should.BeTrue)
	a.So(qos, should.Equal, 2)

	matches := s.Matches("foo")
	a.So(matches, should.Contain, "foo")
	a.So(matches, should.Contain, "+")

	a.So(s.Remove("foo"), should.BeTrue)
	a.So(s.Remove("+"), should.BeTrue)

	a.So(s.Remove("other"), should.BeFalse)

	_, ok = s.Match("foo")
	a.So(ok, should.BeFalse)

	s.Clear()
	a.So(s.Subscriptions(), should.BeEmpty)
	a.So(s.SubscriptionTopics(), should.BeEmpty)
}
