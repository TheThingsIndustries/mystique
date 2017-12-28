// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package mystique implements an MQTT server.
// See the cmd package for the main executables.
package mystique

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/TheThingsIndustries/mystique/pkg/apex"
	"github.com/TheThingsIndustries/mystique/pkg/inspect"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/server"
	"github.com/prometheus/client_golang/prometheus"
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
	pflag.String("listen.http", ":1880", "TCP address for HTTP+websocket server to listen on")
	pflag.String("listen.https", ":1443", "TLS address for HTTP+websocket server to listen on")
	pflag.String("websocket.pattern", "/mqtt", "URL pattern for websocket server to be registered on")
	pflag.String("listen.status", ":9383", "Address for status server to listen on")
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

var certificateExpiry = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "tls",
	Name:      "certificate_expiry_seconds",
	Help:      "Expiry date of the TLS certificate.",
}, []string{"fingerprint"})

func init() {
	prometheus.MustRegister(certificateExpiry)
}

// RunServer the server
func RunServer(s server.Server) {
	wss := mqttnet.Websocket(s.Handle)

	var tlsConfig *tls.Config
	certFile, keyFile := viper.GetString("tls.cert"), viper.GetString("tls.key")
	if certFile != "" && keyFile != "" {
		certPEMBlock, err := ioutil.ReadFile(certFile)
		if err != nil {
			logger.WithError(err).Fatal("Could not read TLS certificate")
		}
		keyPEMBlock, err := ioutil.ReadFile(keyFile)
		if err != nil {
			logger.WithError(err).Fatal("Could not read TLS private key")
		}
		cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		if err != nil {
			logger.WithError(err).Fatal("Could not build X509 keypair")
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		if p, _ := pem.Decode(certPEMBlock); p != nil && p.Type == "CERTIFICATE" {
			if cert, err := x509.ParseCertificate(p.Bytes); err == nil {
				sum := sha1.Sum(cert.Raw)
				certificateExpiry.WithLabelValues(hex.EncodeToString(sum[:])).Set(float64(cert.NotAfter.Unix()))
			}
		}
	}

	if listen := viper.GetString("listen.status"); listen != "" {
		http.Handle("/mqtt", wss)
		http.Handle("/metrics", promhttp.Handler())
		http.Handle("/debug/sessions", inspect.Sessions(s))
		http.Handle("/debug/retained", inspect.Retained(s))
		logger.WithField("address", listen).Info("Starting status+debug+metrics server")
		go func() {
			err := http.ListenAndServe(listen, nil)
			if err != nil {
				logger.WithError(err).Fatal("Could not start status+debug+metrics server")
			}
		}()
	}

	if listen := viper.GetString("listen.tcp"); listen != "" {
		logger.WithField("address", listen).Info("Starting MQTT server")
		lis, err := mqttnet.Listen("tcp", listen)
		if err != nil {
			logger.WithError(err).Fatal("Could not start MQTT server")
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
		if tlsConfig != nil {
			logger.WithField("address", listen).Info("Starting MQTT+TLS server")
			tlsLis, err := tls.Listen("tcp", listen, tlsConfig)
			if err != nil {
				logger.WithError(err).Fatal("Could not start MQTT+TLS server")
			}
			defer tlsLis.Close()

			lis := mqttnet.NewListener(tlsLis, "tls")

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

	if _, err := os.Stat("example/websocket_client.html"); err == nil {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "example/websocket_client.html")
		})
	}

	if listen := viper.GetString("listen.http"); listen != "" {
		logger.WithField("address", listen).Info("Starting HTTP+ws server")
		lis, err := net.Listen("tcp", listen)
		if err != nil {
			logger.WithError(err).Fatal("Could not start HTTP+ws server")
		}
		defer lis.Close()

		go func() {
			err := http.Serve(lis, mux)
			if err != nil {
				logger.WithError(err).Error("Could not serve HTTP+ws")
			}
		}()
	}

	if listen := viper.GetString("listen.https"); listen != "" {
		if tlsConfig != nil {
			logger.WithField("address", listen).Info("Starting HTTPS+wss server")
			tlsLis, err := tls.Listen("tcp", listen, tlsConfig)
			if err != nil {
				logger.WithError(err).Fatal("Could not start HTTPS+wss server")
			}
			defer tlsLis.Close()

			go func() {
				err := http.Serve(tlsLis, mux)
				if err != nil {
					logger.WithError(err).Error("Could not serve HTTPS+wss")
				}
			}()
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	signal := (<-sigChan).String()
	logger.WithField("signal", signal).Info("Signal received")
}
