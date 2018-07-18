package bridge

import (
	"context"
	"strings"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/ttnv2"
	"github.com/TheThingsIndustries/mystique/pkg/ttnv2/discovery"
	"github.com/TheThingsIndustries/mystique/pkg/ttnv2/router"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func (b *Bridge) Connect(discoveryServer, routerID string) error {
	ctx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
	defer cancel()

	b.logger.WithField("discovery_address", discoveryServer).Info("Connecting to Discovery Server...")
	dscConn, err := grpc.DialContext(ctx,
		discoveryServer,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			grpc_prometheus.StreamClientInterceptor,
		)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_prometheus.UnaryClientInterceptor,
		)),
	)
	if err != nil {
		return err
	}
	defer dscConn.Close()
	b.logger.Info("Connected to Discovery Server")

	b.logger.WithField("router_id", routerID).Info("Getting Router Announcement...")
	dscClient := discovery.NewDiscoveryClient(dscConn)
	announcement, err := dscClient.Get(ctx, &ttnv2.GetRequest{
		ServiceName: "router",
		ID:          routerID,
	})
	if err != nil {
		return err
	}
	b.logger.Info("Got Router Announcement")
	routerAddr := strings.Split(announcement.NetAddress, ",")[0]
	b.logger.WithField("router_address", routerAddr).Info("Connecting to Router...")
	tlsConfig, err := announcement.GetTLSConfig()
	if err != nil {
		return err
	}
	b.routerConn, err = grpc.DialContext(ctx,
		routerAddr,
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Minute,
			PermitWithoutStream: true,
		}),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			grpc_prometheus.StreamClientInterceptor,
		)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_prometheus.UnaryClientInterceptor,
		)),
	)
	if err != nil {
		return err
	}
	b.logger.Info("Connected to Router")

	b.router = router.NewRouterClient(b.routerConn)

	return nil
}
