// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package mystique implements an MQTT server.
// See the cmd package for the main executables.
package mystique

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/TheThingsIndustries/mystique/pkg/apex"
	"github.com/TheThingsIndustries/mystique/pkg/inspect"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	ctx        = context.Background()
	logger     = apex.Log
	configured = false

	s server.Server
)

// Context returns the global context
func Context() context.Context {
	if !configured {
		panic("mystique.Configure() was not called")
	}
	return ctx
}

// Configure the binary
func Configure(binaryName string) {
	pflag.BoolP("debug", "d", false, "Print debug logs")
	pflag.String("listen.tcp", ":1883", "TCP address for MQTT server to listen on")
	pflag.String("listen.tls", ":8883", "TLS address for MQTT server to listen on")
	pflag.String("listen.http", ":8080", "TCP address for HTTP+websocket server to listen on")
	pflag.String("listen.https", ":8443", "TLS address for HTTP+websocket server to listen on")
	pflag.String("websocket.pattern", "/mqtt", "URL pattern for websocket server to be registered on")
	pflag.String("listen.status", ":6060", "Address for status server to listen on")
	pflag.String("tls.cert", "", "Location of the TLS certificate")
	pflag.String("tls.key", "", "Location of the TLS key")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", binaryName)
		fmt.Fprintln(os.Stderr, "Options:")
		pflag.PrintDefaults()
	}

	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if viper.GetBool("debug") {
		apex.SetLevelFromString("debug")
	}
	ctx = log.NewContext(ctx, logger)

	configured = true
}

// RunServer the server
func RunServer(s server.Server) {
	wss := mqttnet.Websocket(s.Handle)

	if listenStatus := viper.GetString("listen.status"); listenStatus != "" {
		http.Handle("/mqtt", wss)
		http.Handle("/metrics", promhttp.Handler())
		http.Handle("/inspect/sessions", inspect.Sessions(s))
		go http.ListenAndServe(listenStatus, nil)
	}

	if listen := viper.GetString("listen.tcp"); listen != "" {
		logger.WithField("address", listen).Info("Starting MQTT server")
		lis, err := mqttnet.Listen("tcp", listen)
		if err != nil {
			logger.WithError(err).Fatal("Could not start server")
		}
		defer lis.Close()

		go func() {
			for {
				conn, err := lis.Accept()
				if err != nil {
					logger.WithError(err).Error("Could not accept connection")
					return
				}
				go s.Handle(conn)
			}
		}()
	}

	if listen := viper.GetString("listen.tls"); listen != "" {
		certFile, keyFile := viper.GetString("tls.cert"), viper.GetString("tls.key")
		if certFile != "" && keyFile != "" {
			cert, err := tls.LoadX509KeyPair(filepath.Clean(certFile), filepath.Clean(keyFile))
			if err != nil {
				logger.WithError(err).Fatal("Could not read TLS keypair")
			}

			logger.WithField("address", listen).Info("Starting MQTT+TLS server")
			tlsLis, err := tls.Listen("tcp", listen, &tls.Config{
				Certificates: []tls.Certificate{cert},
			})
			if err != nil {
				logger.WithError(err).Fatal("Could not start server")
			}
			defer tlsLis.Close()

			lis := mqttnet.NewListener(tlsLis)

			go func() {
				for {
					conn, err := lis.Accept()
					if err != nil {
						logger.WithError(err).Error("Could not accept connection")
						return
					}
					go s.Handle(conn)
				}
			}()
		}
	}

	mux := http.NewServeMux()
	mux.Handle(viper.GetString("websocket.pattern"), wss)

	if listen := viper.GetString("listen.http"); listen != "" {
		logger.WithField("address", listen).Info("Starting HTTP server")
		lis, err := net.Listen("tcp", listen)
		if err != nil {
			logger.WithError(err).Fatal("Could not start server")
		}
		defer lis.Close()

		go func() {
			err := http.Serve(lis, mux)
			if err != nil {
				logger.WithError(err).Error("Could not serve HTTP")
			}
		}()
	}

	if listen := viper.GetString("listen.https"); listen != "" {
		certFile, keyFile := viper.GetString("tls.cert"), viper.GetString("tls.key")
		if certFile != "" && keyFile != "" {
			cert, err := tls.LoadX509KeyPair(filepath.Clean(certFile), filepath.Clean(keyFile))
			if err != nil {
				logger.WithError(err).Fatal("Could not read TLS keypair")
			}

			logger.WithField("address", listen).Info("Starting HTTPS server")
			tlsLis, err := tls.Listen("tcp", listen, &tls.Config{
				Certificates: []tls.Certificate{cert},
			})
			if err != nil {
				logger.WithError(err).Fatal("Could not start server")
			}
			defer tlsLis.Close()

			go func() {
				err := http.Serve(tlsLis, mux)
				if err != nil {
					logger.WithError(err).Error("Could not serve HTTP")
				}
			}()
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	signal := (<-sigChan).String()
	logger.WithField("signal", signal).Info("Signal received")
}
