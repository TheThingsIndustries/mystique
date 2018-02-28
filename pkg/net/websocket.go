// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package net

import (
	"fmt"
	"net"
	"net/http"

	"golang.org/x/net/websocket"
)

type wsConn struct {
	Conn
	remoteAddr net.Addr
}

func (c *wsConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// Websocket returns an http.Handler that exposes MQTT over websockets.
func Websocket(handle func(Conn)) http.Handler {
	return websocket.Server{
		Handshake: func(config *websocket.Config, req *http.Request) (err error) {
			config.Origin, err = websocket.Origin(config, req)
			if err != nil {
				return err
			}
			if config.Origin == nil {
				return fmt.Errorf("empty origin")
			}
			var selectedProtocol string
			for _, protocol := range config.Protocol {
				switch protocol {
				case "mqtt", "mqttv3.1":
					selectedProtocol = protocol
					break
				}
			}
			if selectedProtocol == "" {
				return fmt.Errorf("no suitable subprotocol")
			}
			return nil
		},
		Handler: func(ws *websocket.Conn) {
			ws.PayloadType = websocket.BinaryFrame
			var (
				addr net.Addr
				err  error
			)
			addr, err = net.ResolveTCPAddr("tcp", ws.Request().RemoteAddr)
			if err != nil {
				addr = ws.RemoteAddr()
			}
			transport := "ws"
			if ws.Config().TlsConfig != nil {
				transport = "wss"
			}
			conn := &wsConn{
				Conn:       NewConn(ws, transport),
				remoteAddr: addr,
			}
			handle(conn)
		},
	}
}
