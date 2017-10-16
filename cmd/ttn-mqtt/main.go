// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// The TTN MQTT server adds TTN Authentication to the basic Mystique MQTT Server.
//
// Usage: ttn-mqtt [options]
//
// Options:
//         --auth.handler.password string          Handler password (leave empty to disable user)
//         --auth.handler.username string          Handler username (default "handler")
//         --auth.root.password string             Root password (leave empty to disable user)
//         --auth.root.username string             Root username (default "root")
//         --auth.router.password string           Router password (leave empty to disable user)
//         --auth.router.username string           Router username (default "router")
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
	"regexp"
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

	pflag.String("auth.router.username", "router", "Router username")
	pflag.String("auth.router.password", "", "Router password (leave empty to disable user)")

	pflag.String("auth.handler.username", "handler", "Handler username")
	pflag.String("auth.handler.password", "", "Handler password (leave empty to disable user)")

	pflag.StringSlice("auth.ttn.account-server", []string{
		"ttn-account-v2=https://account.thethingsnetwork.org",
	}, "TTN Account Servers")

	mystique.Configure("ttn-mqtt")

	logger := log.FromContext(mystique.Context())

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
			logger.Warn("You are using insecure root credentials, use the --auth.root.username and --auth.root.password arguments to change them")
		}
		auth.AddSuperUser(rootUsername, []byte(rootPassword), ttnauth.Access{Root: true})
	}

	routerUsername, routerPassword := viper.GetString("auth.router.username"), viper.GetString("auth.router.password")
	if routerUsername != "" && routerPassword != "" {
		if routerUsername == routerPassword {
			logger.Warn("You are using insecure router credentials, use the --auth.router.username and --auth.router.password arguments to change them")
		}
		auth.AddSuperUser(routerUsername, []byte(routerPassword), ttnauth.Access{
			Read: []string{"connect", "disconnect"},
			WritePattern: []*regexp.Regexp{
				regexp.MustCompile("^" + ttnauth.IDRegexp + "/down$"),
			},
			ReadPattern: []*regexp.Regexp{
				regexp.MustCompile("^" + ttnauth.IDRegexp + "/up$"),
				regexp.MustCompile("^" + ttnauth.IDRegexp + "/status$"),
			},
		})
	}

	handlerUsername, handlerPassword := viper.GetString("auth.handler.username"), viper.GetString("auth.handler.password")
	if handlerUsername != "" && handlerPassword != "" {
		if handlerUsername == handlerPassword {
			logger.Warn("You are using insecure handler credentials, use the --auth.handler.username and --auth.handler.password arguments to change them")
		}
		auth.AddSuperUser(handlerUsername, []byte(handlerPassword), ttnauth.Access{
			Read: []string{"connect", "disconnect"},
			WritePattern: []*regexp.Regexp{
				regexp.MustCompile("^" + ttnauth.IDRegexp + "/devices" + ttnauth.IDRegexp + "/up$"),
				regexp.MustCompile("^" + ttnauth.IDRegexp + "/devices" + ttnauth.IDRegexp + "/events$"),
				regexp.MustCompile("^" + ttnauth.IDRegexp + "/events$"),
			},
			ReadPattern: []*regexp.Regexp{
				regexp.MustCompile("^" + ttnauth.IDRegexp + "/devices" + ttnauth.IDRegexp + "/down$"),
			},
		})
	}

	s := server.New(mystique.Context(), server.WithAuth(auth))

	mystique.RunServer(s)
}
