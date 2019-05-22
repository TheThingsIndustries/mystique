package bridge

import (
	"context"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/session"
	"github.com/TheThingsIndustries/mystique/pkg/ttnv2"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func (l *link) runStreams() (err error) {
	ctx, cancel := context.WithCancel(l.ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	var (
		uplinkErrCh   = make(chan error, 1)
		downlinkErrCh = make(chan error, 1)
		statusErrCh   = make(chan error, 1)
	)

	logger := log.FromContext(l.ctx)

	logger.Debug("Starting Uplink stream")
	uplink, err := l.bridge.router.Uplink(ctx, grpc.PerRPCCredentials(l))
	if err != nil {
		return err
	}
	if _, err = uplink.Header(); err != nil {
		return err
	}
	go func() {
		empty := new(types.Empty)
		uplinkErr := uplink.RecvMsg(empty)
		uplinkErrCh <- uplinkErr
	}()

	logger.Debug("Starting downlink stream")
	downlink, err := l.bridge.router.Subscribe(ctx, &types.Empty{}, grpc.PerRPCCredentials(l))
	if err != nil {
		return err
	}
	if _, err = downlink.Header(); err != nil {
		return err
	}
	go func() {
		for {
			msg, downlinkErr := downlink.Recv()
			if downlinkErr != nil {
				downlinkErrCh <- downlinkErr
				break
			}
			l.downlinkCh <- msg
		}
	}()

	logger.Debug("Starting status stream")
	status, err := l.bridge.router.GatewayStatus(ctx, grpc.PerRPCCredentials(l))
	if err != nil {
		return err
	}
	if _, err = status.Header(); err != nil {
		return err
	}
	go func() {
		empty := new(types.Empty)
		statusErr := status.RecvMsg(empty)
		statusErrCh <- statusErr
	}()

	l.streamsMu.Lock()

	l.uplinkStream = uplink
	l.downlinkStream = downlink
	l.statusStream = status

	if l.streamCancel != nil {
		l.streamCancel()
	}
	l.streamCancel = cancel

	l.streamsMu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-uplinkErrCh:
			if grpc.Code(err) == codes.Canceled {
				return context.Canceled
			}
			return err
		case err := <-downlinkErrCh:
			if grpc.Code(err) == codes.Canceled {
				return context.Canceled
			}
			return err
		case err := <-statusErrCh:
			if grpc.Code(err) == codes.Canceled {
				return context.Canceled
			}
			return err
		}
	}
}

// start a new link -- NON-BLOCKING
func (b *Bridge) startLink(s session.Session) *link {
	logger := b.logger.WithField("gateway_id", s.AuthInfo().Username)

	l := &link{
		bridge:     b,
		ctx:        log.NewContext(b.ctx, logger),
		gatewayID:  s.AuthInfo().Username,
		gatewayKey: string(s.AuthInfo().Password),
		downlinkCh: make(chan *ttnv2.DownlinkMessage),
	}
	l.ctx, l.cancel = context.WithCancel(l.ctx) // not s.Context(), because its lifetime is more limited

	go func() {
		for {
			select {
			case <-l.ctx.Done():
				return
			case msg := <-l.downlinkCh:
				msg.Trace = nil
				b.SendDown(l.gatewayID, msg)
			}
		}
	}()

	go func() {
		gtw, err := l.GetGateway()
		if err != nil {
			b.logger.WithError(err).Warn("Could not get token for gateway")
			return
		}
		logger := logger.WithField("gateway_owner", gtw.Owner.Username)
		logger.Info("Link gateway to Router")
		defer logger.Info("Unlink gateway from Router")

		for {
			if l.ctx.Err() != nil {
				return
			}
			err = l.runStreams()
			if err != nil {
				if l.ctx.Err() != nil {
					return
				}
				logger.WithError(err).Warn("Error in gateway link")
				if grpc.Code(err) != codes.Unavailable {
					time.Sleep(2 * time.Second) // TODO: backoff
				} else {
					time.Sleep(10 * time.Second) // TODO: backoff
				}
			}
		}
	}()

	return l
}

func (l *link) SendUp(msg proto.Message) {
	logger := log.FromContext(l.ctx)
	l.streamsMu.RLock()
	uplink, status := l.uplinkStream, l.statusStream
	l.streamsMu.RUnlock()

	gateway, _ := l.GetGateway()

	switch msg := msg.(type) {
	case *ttnv2.UplinkMessage:
		if uplink == nil {
			logger.Warn("Not ready to send uplink message")
			return
		}
		ProcessUplink(gateway, msg)
		if err := uplink.Send(msg); err != nil {
			logger.WithError(err).Warn("Could not send uplink message")
		} else {
			logger.Info("Sent uplink message")
		}
	case *ttnv2.StatusMessage:
		if status == nil {
			logger.Warn("Not ready to send status message")
			return
		}
		ProcessStatus(gateway, msg)
		if err := status.Send(msg); err != nil {
			logger.WithError(err).Warn("Could not send status message")
		} else {
			logger.Info("Sent status message")
		}
	}
}
