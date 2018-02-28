// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package topic

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMatch(t *testing.T) {
	a := assertions.New(t)
	a.So(Match("a", "b"), should.BeFalse)
	a.So(Match("/", "/"), should.BeTrue)
	a.So(Match("//", "//"), should.BeTrue)
	a.So(Match("a/b/c", "a/b/c"), should.BeTrue)
	a.So(Match("a/b/c", "a/b"), should.BeFalse)
	a.So(Match("a/b", "a/b/c"), should.BeFalse)
	a.So(Match("a", "+"), should.BeTrue)
	a.So(Match("/", "+/+"), should.BeTrue)
	a.So(Match("a", "#"), should.BeTrue)
	a.So(Match("/", "#"), should.BeTrue)
	a.So(Match("a/b", "+/a"), should.BeFalse)
	a.So(Match("a/b", "+/b"), should.BeTrue)
	a.So(Match("a/b", "a/+"), should.BeTrue)
	a.So(Match("a/b/c", "a/+"), should.BeFalse)
	a.So(Match("a/b/c", "a/#"), should.BeTrue)
	a.So(Match("$SYS/number", "#"), should.BeFalse)
	a.So(Match("$SYS/number", "+/number"), should.BeFalse)
	a.So(Match("$SYS/number", "$SYS/#"), should.BeTrue)
	a.So(Match("$SYS/number", "$SYS/+"), should.BeTrue)
}

func TestValidate(t *testing.T) {
	a := assertions.New(t)
	a.So(ValidateTopic(""), should.NotBeNil)
	a.So(ValidateFilter(""), should.NotBeNil)
	a.So(ValidateTopic("foo bar baz/subTopïç⚡"), should.BeNil)
	a.So(ValidateFilter("foo bar baz/subTopïç⚡"), should.BeNil)
	a.So(ValidateFilter("hello\u0000world"), should.NotBeNil)
	a.So(ValidateTopic("a/+"), should.NotBeNil)
	a.So(ValidateFilter("a/+"), should.BeNil)
	a.So(ValidateTopic("a/#"), should.NotBeNil)
	a.So(ValidateFilter("a/#"), should.BeNil)
	a.So(ValidateTopic("a+"), should.NotBeNil)
	a.So(ValidateFilter("a+"), should.NotBeNil)
	a.So(ValidateTopic("a#"), should.NotBeNil)
	a.So(ValidateFilter("a#"), should.NotBeNil)
}
