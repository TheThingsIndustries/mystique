// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package log

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestContext(t *testing.T) {
	a := assertions.New(t)

	a.So(FromContext(context.Background()), should.Equal, Noop)

	logger := &noop{}
	ctx := NewContext(context.Background(), logger)
	a.So(FromContext(ctx), should.Equal, logger)
}

func TestFields(t *testing.T) {
	a := assertions.New(t)
	fields := map[string]interface{}{
		"foo": "foo",
	}
	a.So(Fields(fields).Fields(), should.Equal, fields)
}

func TestNoop(t *testing.T) {
	a := assertions.New(t)
	noop := &noop{}
	a.So(func() { noop.Debug("") }, should.NotPanic)
	a.So(func() { noop.Info("") }, should.NotPanic)
	a.So(func() { noop.Warn("") }, should.NotPanic)
	a.So(func() { noop.Error("") }, should.NotPanic)
	a.So(func() { noop.Fatal("") }, should.NotPanic)
	a.So(func() { noop.Debugf("") }, should.NotPanic)
	a.So(func() { noop.Infof("") }, should.NotPanic)
	a.So(func() { noop.Warnf("") }, should.NotPanic)
	a.So(func() { noop.Errorf("") }, should.NotPanic)
	a.So(func() { noop.Fatalf("") }, should.NotPanic)
	a.So(noop.WithField("foo", "foo"), should.Equal, noop)
	a.So(noop.WithFields(Fields(map[string]interface{}{})), should.Equal, noop)
	a.So(noop.WithError(errors.New("foo")), should.Equal, noop)
}
