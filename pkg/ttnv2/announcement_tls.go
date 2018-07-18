package ttnv2

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strings"
)

// GetTLSConfig returns the TLS config for the announcement
func (a *Announcement) GetTLSConfig() (*tls.Config, error) {
	if a.Certificate == "" {
		return nil, nil
	}
	if a.NetAddress == "" {
		return nil, errors.New("No address known for this component")
	}
	netAddress := strings.Split(a.NetAddress, ",")[0]
	netHost, _, _ := net.SplitHostPort(netAddress)
	rootCAs := x509.NewCertPool()
	ok := rootCAs.AppendCertsFromPEM([]byte(a.Certificate))
	if !ok {
		return nil, errors.New("Could not read component certificate")
	}
	return &tls.Config{ServerName: netHost, RootCAs: rootCAs}, nil
}
