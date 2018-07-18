package bridge

import (
	"context"
	"sync"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/ttnv2"
	"github.com/TheThingsIndustries/mystique/pkg/ttnv2/router"
)

type link struct {
	bridge *Bridge

	connections uint // only accessed by the bridge

	ctx    context.Context
	cancel context.CancelFunc

	gatewayID  string
	gatewayKey string

	gatewayMu           sync.Mutex
	gateway             *Gateway
	gatewayUpdated      time.Time
	gatewayTokenExpires time.Time

	downlinkCh chan *ttnv2.DownlinkMessage

	streamsMu      sync.RWMutex
	streamCancel   context.CancelFunc
	uplinkStream   router.Router_UplinkClient
	downlinkStream router.Router_SubscribeClient
	statusStream   router.Router_GatewayStatusClient
}

func (l *link) GetGateway() (*Gateway, error) {
	l.gatewayMu.Lock()
	defer l.gatewayMu.Unlock()
	if l.gateway != nil && time.Since(l.gatewayUpdated) < time.Hour && time.Until(l.gatewayTokenExpires) > time.Hour {
		return l.gateway, nil
	}
	gateway, err := l.bridge.fetchGateway(l.gatewayID, l.gatewayKey)
	if err != nil {
		return nil, err
	}
	l.gateway = gateway
	l.gatewayUpdated = time.Now()
	l.gatewayTokenExpires = l.gatewayUpdated.Add(time.Duration(gateway.Token.ExpiresIn) * time.Second)
	return gateway, nil
}

// GetRequestMetadata implements credentials.PerRPCCredentials
func (l *link) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	gateway, err := l.GetGateway()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"id":    l.gatewayID,
		"token": gateway.Token.AccessToken,
	}, nil
}

// RequireTransportSecurity implements credentials.PerRPCCredentials
func (l *link) RequireTransportSecurity() bool {
	return true
}

func (l *link) Close() error {
	l.cancel()
	return nil
}
