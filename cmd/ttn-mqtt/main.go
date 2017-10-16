// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// The TTN MQTT server adds TTN Authentication to the basic Mystique MQTT Server.
//
// Usage: ttn-mqtt [options]
//
// Options:
//         --auth.root.password string             Root password (leave empty to disable user)
//         --auth.root.username string             Root username (default "root")
//         --auth.ttn.account-server stringSlice   TTN Account Servers (default [ttn-account-v2=https://account.thethingsnetwork.org])
//     -d, --debug                                 Print debug logs
//         --listen.status string                  Address for status server to listen on (default ":6060")
//         --listen.tcp string                     TCP address for MQTT server to listen on (default ":1883")
//         --listen.tls string                     TLS address for MQTT server to listen on (default ":8883")
//         --tls.cert string                       Location of the TLS certificate
//         --tls.key string                        Location of the TLS key
package main

import (
	_ "net/http/pprof" // Add pprof handlers to the default http mux
	"strings"

	"github.com/TheThingsIndustries/mystique"
	"github.com/TheThingsIndustries/mystique/pkg/auth/ttnauth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/server"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.String("auth.root.username", "root", "Root username")
	pflag.String("auth.root.password", "", "Root password (leave empty to disable user)")

	pflag.StringSlice("auth.ttn.account-server", []string{
		"ttn-account-v2=https://account.thethingsnetwork.org",
	}, "TTN Account Servers")

	mystique.Configure("ttn-mqtt")

	serverSlice := viper.GetStringSlice("auth.ttn.account-server")
	accountServers := make(map[string]string, len(serverSlice))
	for _, server := range serverSlice {
		parts := strings.SplitN(server, "=", 2)
		if len(parts) != 2 {
			continue
		}
		accountServers[parts[0]] = parts[1]
	}

	auth := ttnauth.New(accountServers)

	rootUsername, rootPassword := viper.GetString("auth.root.username"), viper.GetString("auth.root.password")
	if rootUsername != "" && rootPassword != "" {
		if rootUsername == rootPassword {
			log.FromContext(mystique.Context()).Warn("You are using insecure root credentials, use the --auth.root.username and --auth.root.password arguments to change them")
		}
		auth.AddSuperUser(rootUsername, []byte(rootPassword), ttnauth.Access{Root: true})
	}

	s := server.New(mystique.Context(), server.WithAuth(auth))

	mystique.RunServer(s)
}
