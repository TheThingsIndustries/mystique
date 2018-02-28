// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// The Mystique MQTT Server is a simple MQTT server.
//
//     Usage: mystique-server [options]
//
//     Options:
//     -d, --debug                      Print debug logs
//         --listen.http string         TCP address for HTTP+websocket server to listen on (default ":1880")
//         --listen.https string        TLS address for HTTP+websocket server to listen on (default ":1443")
//         --listen.status string       Address for status server to listen on (default ":9383")
//         --listen.tcp string          TCP address for MQTT server to listen on (default ":1883")
//         --listen.tls string          TLS address for MQTT server to listen on (default ":8883")
//         --tls.cert string            Location of the TLS certificate
//         --tls.key string             Location of the TLS key
//         --websocket.pattern string   URL pattern for websocket server to be registered on (default "/mqtt")
package main

import (
	_ "net/http/pprof" // Add pprof handlers to the default http mux

	"github.com/TheThingsIndustries/mystique"
	"github.com/TheThingsIndustries/mystique/pkg/server"
)

func main() {
	mystique.Configure("mystique-server")
	s := server.New(mystique.Context())
	mystique.RunServer(s)
}
