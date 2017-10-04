// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// The Mystique MQTT Server is a simple MQTT server.
//
//     Usage: mystique-server [options]
//
//     Options:
//     -d, --debug                  Print debug logs
//         --listen.status string   Address for status server to listen on (default ":6060")
//         --listen.tcp string      TCP address for MQTT server to listen on (default ":1883")
//         --listen.tls string      TLS address for MQTT server to listen on (default ":8883")
//         --tls.cert string        Location of the TLS certificate
//         --tls.key string         Location of the TLS key
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
