// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package auth

import (
	"errors"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type alwaysAuth struct {
	ok bool
}

func (a alwaysAuth) Connect(info *Info) (err error) {
	if !a.ok {
		err = errors.New("computer says no")
	}
	return
}
func (a alwaysAuth) Subscribe(info *Info, requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error) {
	if !a.ok {
		err = errors.New("computer says no")
	}
	return
}
func (a alwaysAuth) CanRead(info *Info, topic ...string) bool {
	return a.ok
}
func (a alwaysAuth) CanWrite(info *Info, topic ...string) bool {
	return a.ok
}

func TestAuth(t *testing.T) {
	a := assertions.New(t)

	var i *Info // nil

	_, _, err := i.Subscribe("topic", 0)
	a.So(err, should.NotBeNil)
	a.So(i.CanRead("topic"), should.BeFalse)
	a.So(i.CanWrite("topic"), should.BeFalse)

	i = &Info{}

	_, _, err = i.Subscribe("topic", 0)
	a.So(err, should.BeNil)
	a.So(i.CanRead("topic"), should.BeTrue)
	a.So(i.CanWrite("topic"), should.BeTrue)

	i.Interface = alwaysAuth{false}

	_, _, err = i.Subscribe("topic", 0)
	a.So(err, should.NotBeNil)
	a.So(i.CanRead("topic"), should.BeFalse)
	a.So(i.CanWrite("topic"), should.BeFalse)

	i.Interface = alwaysAuth{true}

	_, _, err = i.Subscribe("topic", 0)
	a.So(err, should.BeNil)
	a.So(i.CanRead("topic"), should.BeTrue)
	a.So(i.CanWrite("topic"), should.BeTrue)
}
