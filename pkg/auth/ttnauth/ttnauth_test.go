// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package ttnauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestTTNAuth(t *testing.T) {
	a := assertions.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/api/v2/applications/no/rights":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`null`))
		case "/api/v2/applications/test/rights":
			switch r.Header.Get("Authorization") {
			case "Key test.app":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`["messages:up:r","messages:down:w"]`))
			default:
				w.WriteHeader(http.StatusUnauthorized)
			}
		case "/api/v2/gateways/no/rights":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`null`))
		case "/api/v2/gateways/test/rights":
			switch r.Header.Get("Authorization") {
			case "Key test.gtw":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`["rights go here"]`))
			default:
				w.WriteHeader(http.StatusUnauthorized)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	s := New(map[string]string{"test": ts.URL})
	s.AuthenticateGateways()
	s.AuthenticateApplications()
	s.AddSuperUser("root", []byte("rootpass"), Access{Root: true})

	a.So(s.Connect(&auth.Info{}), should.NotBeNil)

	empty := &auth.Info{
		Interface: &TTNAuth{},
	}

	a.So(empty.CanRead("any"), should.BeFalse)
	a.So(empty.CanWrite("any"), should.BeFalse)
	{
		_, _, err := empty.Subscribe("any", 0)
		a.So(err, should.NotBeNil)
	}

	root := &auth.Info{
		Username: "root",
		Password: []byte("rootpass"),
	}
	a.So(s.Connect(root), should.BeNil)

	a.So(root.CanRead("any"), should.BeTrue)
	a.So(root.CanWrite("any"), should.BeTrue)
	a.So(root.CanRead("$SYS/#"), should.BeTrue)
	a.So(root.CanWrite("$SYS/#"), should.BeFalse)

	incorrect := &auth.Info{
		Username: "test",
		Password: []byte("test.incorrect"),
	}
	a.So(s.Connect(incorrect), should.NotBeNil)

	a.So(s.Connect(&auth.Info{
		Username: "no",
		Password: []byte("test.app"),
	}), should.NotBeNil)

	a.So(s.Connect(&auth.Info{
		Username: "no",
		Password: []byte("test.gtw"),
	}), should.NotBeNil)

	app := &auth.Info{
		Username: "test",
		Password: []byte("test.app"),
	}
	a.So(s.Connect(app), should.BeNil)

	a.So(app.CanRead("test/devices/test/up"), should.BeTrue)
	a.So(app.CanRead("other/devices/test/up"), should.BeFalse)
	a.So(app.CanRead("test/devices/test/up/temperature"), should.BeTrue)
	a.So(app.CanRead("other/devices/test/up/temperature"), should.BeFalse)
	a.So(app.CanRead("test/devices/test/events"), should.BeTrue)
	a.So(app.CanRead("other/devices/test/events"), should.BeFalse)
	a.So(app.CanRead("test/devices/test/events/downlink"), should.BeTrue)
	a.So(app.CanRead("other/devices/test/events/downlink"), should.BeFalse)
	a.So(app.CanWrite("test/devices/test/down"), should.BeTrue)
	a.So(app.CanWrite("other/devices/test/down"), should.BeFalse)
	a.So(app.CanRead("test/events"), should.BeTrue)
	a.So(app.CanRead("other/events"), should.BeFalse)
	a.So(app.CanRead("test/events/activate"), should.BeTrue)
	a.So(app.CanRead("other/events/activate"), should.BeFalse)
	a.So(app.CanRead("$SYS/#"), should.BeFalse)

	{
		_, _, err := app.Subscribe("root", 0)
		a.So(err, should.NotBeNil)

		topic, _, err := app.Subscribe("#", 0)
		a.So(err, should.BeNil)
		a.So(topic, should.Equal, "test/#")
	}

	gtw := &auth.Info{
		Username: "test",
		Password: []byte("test.gtw"),
	}
	a.So(s.Connect(gtw), should.BeNil)

	a.So(gtw.CanRead("connect"), should.BeFalse)
	a.So(gtw.CanWrite("connect"), should.BeTrue)

	a.So(gtw.CanRead("test/down"), should.BeTrue)
	a.So(gtw.CanRead("other/down"), should.BeFalse)
	a.So(gtw.CanWrite("test/up"), should.BeTrue)
	a.So(gtw.CanWrite("other/up"), should.BeFalse)
	a.So(gtw.CanRead("$SYS/#"), should.BeFalse)

	{
		_, _, err := gtw.Subscribe("connect", 0)
		a.So(err, should.NotBeNil)

		topic, _, err := app.Subscribe("#", 0)
		a.So(err, should.BeNil)
		a.So(topic, should.Equal, "test/#")

		topic, _, err = app.Subscribe("+/up", 0)
		a.So(err, should.BeNil)
		a.So(topic, should.Equal, "test/up")
	}

	hdl := &auth.Info{
		Interface: s,
		Username:  "$handler",
		Metadata:  &HandlerAccess,
	}

	a.So(hdl.CanWrite("test/devices/test/up"), should.BeTrue)
	a.So(hdl.CanWrite("test/devices/test/up/temperature"), should.BeTrue)
	a.So(hdl.CanWrite("test/devices/test/events"), should.BeTrue)
	a.So(hdl.CanWrite("test/devices/test/events/downlink"), should.BeTrue)
	a.So(hdl.CanRead("test/devices/test/down"), should.BeTrue)
	a.So(hdl.CanWrite("test/events"), should.BeTrue)
	a.So(hdl.CanWrite("test/events/activate"), should.BeTrue)
	a.So(hdl.CanRead("$SYS/#"), should.BeFalse)
	a.So(hdl.CanWrite("$SYS/#"), should.BeFalse)

	rtr := &auth.Info{
		Interface: s,
		Username:  "$router",
		Metadata:  &RouterAccess,
	}

	a.So(rtr.CanWrite("connect"), should.BeFalse)
	a.So(rtr.CanRead("connect"), should.BeTrue)
	a.So(rtr.CanWrite("test/down"), should.BeTrue)
	a.So(rtr.CanRead("test/up"), should.BeTrue)
	a.So(rtr.CanRead("$SYS/#"), should.BeFalse)
	a.So(rtr.CanWrite("$SYS/#"), should.BeFalse)
}
