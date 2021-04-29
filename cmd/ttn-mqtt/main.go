// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// The TTN MQTT server adds TTN Authentication to the basic Mystique MQTT Server.
//
// Usage: ttn-mqtt [options]
//
// Options:
//         --auth.applications                     Authenticate Applications (default true)
//         --auth.gateways                         Authenticate Gateways (default true)
//         --auth.handler.password string          Handler password (leave empty to disable user)
//         --auth.handler.username string          Handler username (default "$handler")
//         --auth.penalty duration                 Time penalty for a failed login
//         --auth.root.password string             Root password (leave empty to disable user)
//         --auth.root.username string             Root username (default "$root")
//         --auth.router.password string           Router password (leave empty to disable user)
//         --auth.router.username string           Router username (default "$router")
//         --auth.ttn.account-server stringSlice   TTN Account Servers (default [ttn-account-v2=https://account.thethingsnetwork.org])
//     -d, --debug                                 Print debug logs
//         --listen.http string                    TCP address for HTTP+websocket server to listen on (default ":1880")
//         --listen.https string                   TLS address for HTTP+websocket server to listen on (default ":1443")
//         --listen.status string                  Address for status server to listen on (default ":9383")
//         --listen.tcp string                     TCP address for MQTT server to listen on (default ":1883")
//         --listen.tls string                     TLS address for MQTT server to listen on (default ":8883")
//         --tls.cert string                       Location of the TLS certificate
//         --tls.key string                        Location of the TLS key
//         --websocket.pattern string              URL pattern for websocket server to be registered on (default "/mqtt")
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
	"golang.org/x/time/rate"
)

func main() {
	pflag.String("auth.root.username", "$root", "Root username")
	pflag.String("auth.root.password", "", "Root password (leave empty to disable user)")

	pflag.String("auth.router.username", "$router", "Router username")
	pflag.String("auth.router.password", "", "Router password (leave empty to disable user)")

	pflag.Bool("auth.gateways", true, "Authenticate Gateways")

	pflag.String("auth.handler.username", "$handler", "Handler username")
	pflag.String("auth.handler.password", "", "Handler password (leave empty to disable user)")

	pflag.Bool("auth.applications", true, "Authenticate Applications")

	pflag.Duration("auth.penalty", 0, "Time penalty for a failed login")

	pflag.StringSlice("auth.ttn.account-server", []string{
		"ttn-account-v2=https://account.thethingsnetwork.org",
	}, "TTN Account Servers")

	pflag.Int("limit.ip", 0, "Connection limit per IP address")
	pflag.Int("limit.user", 0, "Connection limit per Username")
	pflag.Float64("limit.rate", 10, "Rate limit per connection")

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

	auth.SetLogger(logger)

	ttnIDRegexp := regexp.MustCompile("^" + ttnauth.IDRegexp + "$")

	rootUsername, rootPassword := viper.GetString("auth.root.username"), viper.GetString("auth.root.password")
	if rootUsername != "" && rootPassword != "" {
		if ttnIDRegexp.MatchString(rootUsername) {
			logger.Warnf(`The root username "%s" may clash with TTN usernames, consider prefixing it with $`, rootUsername)
		}
		if rootUsername == rootPassword {
			logger.Warn("The root password equals the username, which is not very secure, use the --auth.root.password flag to change it")
		}
		auth.AddSuperUser(rootUsername, []byte(rootPassword), ttnauth.Access{Root: true})
	}

	routerUsername, routerPassword := viper.GetString("auth.router.username"), viper.GetString("auth.router.password")
	if routerUsername != "" && routerPassword != "" {
		if ttnIDRegexp.MatchString(routerUsername) {
			logger.Warnf(`The router username "%s" may clash with TTN usernames, consider prefixing it with $`, routerUsername)
		}
		if routerUsername == routerPassword {
			logger.Warn("The router password equals the username, which is not very secure, use the --auth.router.password flag to change it")
		}
		auth.AddSuperUser(routerUsername, []byte(routerPassword), ttnauth.RouterAccess)
	}

	if viper.GetBool("auth.gateways") {
		auth.AuthenticateGateways()
	}

	handlerUsername, handlerPassword := viper.GetString("auth.handler.username"), viper.GetString("auth.handler.password")
	if handlerUsername != "" && handlerPassword != "" {
		if ttnIDRegexp.MatchString(handlerUsername) {
			logger.Warnf(`The handler username "%s" may clash with TTN usernames, consider prefixing it with $`, handlerUsername)
		}
		if handlerUsername == handlerPassword {
			logger.Warn("The handler password equals the username, which is not very secure, use the --auth.handler.password flag to change it")
		}
		auth.AddSuperUser(handlerUsername, []byte(handlerPassword), ttnauth.HandlerAccess)
	}

	if viper.GetBool("auth.applications") {
		auth.AuthenticateApplications()
	}

	auth.SetPenalty(viper.GetDuration("auth.penalty"))
	auth.SetRateLimit(rate.Limit(viper.GetFloat64("limit.rate")))

	serverOptions := []server.Option{server.WithAuth(auth)}

	if ipLimit := viper.GetInt("limit.ip"); ipLimit > 0 {
		serverOptions = append(serverOptions, server.WithIPLimits(ipLimit))
	}

	if userLimit := viper.GetInt("limit.user"); userLimit > 0 {
		serverOptions = append(serverOptions, server.WithUserLimits(userLimit))
	}

	s := server.New(mystique.Context(), serverOptions...)

	mystique.RunServer(s)
}
