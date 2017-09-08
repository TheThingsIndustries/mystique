// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"context"
	"testing"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestStore(t *testing.T) {
	a := assertions.New(t)
	ctx := context.Background()
	s := SimpleStore(ctx)

	sess := s.GetOrCreate("foo")
	a.So(sess, should.NotBeNil)

	sess2 := s.GetOrCreate("foo")
	a.So(sess2, should.Equal, sess)

	sess.(*serverSession).authinfo = &auth.Info{} // needed for publish
	s.Publish(&packet.PublishPacket{})

	s.Delete("foo")

	sess3 := s.GetOrCreate("foo")
	a.So(sess3, should.NotEqual, sess)
}
