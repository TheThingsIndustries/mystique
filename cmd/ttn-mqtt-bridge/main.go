// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package main

import (
	_ "net/http/pprof" // Add pprof handlers to the default http mux
	"regexp"
	"strings"

	"github.com/TheThingsIndustries/mystique"
	"github.com/TheThingsIndustries/mystique/cmd/ttn-mqtt-bridge/bridge"
	"github.com/TheThingsIndustries/mystique/pkg/auth/ttnauth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/server"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.String("auth.root.username", "$root", "Root username")
	pflag.String("auth.root.password", "", "Root password (leave empty to disable user)")

	pflag.Duration("auth.penalty", 0, "Time penalty for a failed login")

	pflag.StringSlice("auth.ttn.account-server", []string{
		"ttn-account-v2=https://account.thethingsnetwork.org",
	}, "TTN Account Servers")

	pflag.String("ttn.v2.discovery-server", "discover.thethingsnetwork.org:1900", "TTN v2 Discovery Server")
	pflag.String("ttn.v2.router-id", "ttn-router-eu", "TTN v2 Router ID")

	mystique.Configure("ttn-mqtt-bridge")

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
	bridge.SetGRPCLogger(logger.WithField("namespace", "grpc"))

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

	auth.AuthenticateGateways()

	auth.SetPenalty(viper.GetDuration("auth.penalty"))

	serverOptions := []server.Option{server.WithAuth(auth)}

	bridge := bridge.New(mystique.Context(), accountServers)

	if err := bridge.Connect(viper.GetString("ttn.v2.discovery-server"), viper.GetString("ttn.v2.router-id")); err != nil {
		logger.WithError(err).Fatal("Could not connect to TTN.")
	}

	serverOptions = append(serverOptions, server.WithSessionStore(bridge))

	s := server.New(mystique.Context(), serverOptions...)

	mystique.RunServer(s)
}
