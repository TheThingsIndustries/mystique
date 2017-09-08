// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package net

import (
	"net"
	"testing"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMQTTNet(t *testing.T) {
	a := assertions.New(t)

	lis, err := Listen("tcp", ":0")
	a.So(err, should.BeNil)
	_, port, _ := net.SplitHostPort(lis.Addr().String())
	defer lis.Close()

	var sConn Conn
	sConnWait := make(chan struct{})
	go func() {
		var err error
		sConn, err = lis.Accept()
		a.So(err, should.BeNil)
		close(sConnWait)
	}()

	conn, err := Dial("tcp", net.JoinHostPort("localhost", port))
	a.So(err, should.BeNil)
	defer conn.Close()

	select {
	case <-sConnWait:
	case <-time.After(time.Second):
		t.Error("not connected to server")
		t.Fail()
	}

	a.So(sConn, should.NotBeNil)
	a.So(conn, should.NotBeNil)

	pkt := &packet.ConnectPacket{
		ProtocolName: "MQTT",
	}
	a.So(conn.Send(pkt), should.BeNil)

	recv, err := sConn.Receive()
	a.So(err, should.BeNil)
	a.So(recv, should.Resemble, pkt)
}
