package bridge

import (
	"context"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth/ttnauth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/session"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"github.com/TheThingsIndustries/mystique/pkg/ttnv2"
	"github.com/TheThingsIndustries/mystique/pkg/ttnv2/router"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Bridge struct {
	ctx            context.Context
	logger         log.Interface
	accountServers map[string]string
	http           *http.Client

	routerConn *grpc.ClientConn
	router     router.RouterClient

	store session.Store

	mu    sync.RWMutex
	links map[string]*link
}

func New(ctx context.Context, accountServers map[string]string) *Bridge {
	return &Bridge{
		ctx: metadata.NewOutgoingContext(ctx, metadata.Pairs(
			"service-name", "ttn-mqtt-bridge",
			"service-version", "2.0.x",
		)),
		logger:         log.FromContext(ctx),
		accountServers: accountServers,
		http:           http.DefaultClient,
		store:          session.SimpleStore(),
		links:          make(map[string]*link),
	}
}

func (b *Bridge) Close() error {
	return b.routerConn.Close()
}

// All implements the session.Store interface -- NON-BLOCKING
func (b *Bridge) All() []session.Session { return b.store.All() }

var ttnIDRegexp = regexp.MustCompile("^" + ttnauth.IDRegexp + "$")

// Store implements the session.Store interface -- NON-BLOCKING
func (b *Bridge) Store(s session.Session) {
	authInfo := s.AuthInfo()
	if ttnIDRegexp.MatchString(authInfo.Username) {
		b.logger.WithFields(log.F{
			"gateway_id":  authInfo.Username,
			"remote_addr": authInfo.RemoteAddr,
		}).Infof("Gateway connected to MQTT")
		b.mu.Lock()
		link, ok := b.links[authInfo.Username]
		if !ok {
			link = b.startLink(s)
			b.links[authInfo.Username] = link
		}
		link.connections++
		gatewaysConnected.WithLabelValues(authInfo.Username).Inc()
		b.mu.Unlock()
	}
	b.store.Store(s)
}

// Delete implements the session.Store interface -- NON-BLOCKING
func (b *Bridge) Delete(s session.Session) {
	b.store.Delete(s)
	authInfo := s.AuthInfo()
	if ttnIDRegexp.MatchString(authInfo.Username) {
		b.logger.WithFields(log.F{
			"gateway_id":  authInfo.Username,
			"remote_addr": authInfo.RemoteAddr,
		}).Infof("Gateway disconnected from MQTT")
		b.mu.Lock()
		link, ok := b.links[authInfo.Username]
		if ok {
			link.connections--
			gatewaysDisconnected.WithLabelValues(authInfo.Username).Inc()
			if link.connections == 0 {
				delete(b.links, authInfo.Username)
				link.Close()
			}
		}
		b.mu.Unlock()
	}
}

func (b *Bridge) getLink(username string) *link {
	b.mu.RLock()
	link := b.links[username]
	b.mu.RUnlock()
	return link
}

func (b *Bridge) SendDown(gateway string, msg *ttnv2.DownlinkMessage) {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return
	}
	topicParts := []string{gateway, "down"}
	pkt := &packet.PublishPacket{
		Received:   time.Now(),
		TopicName:  topic.Join(topicParts),
		TopicParts: topicParts,
		Message:    payload,
	}
	b.store.Publish(pkt)
	b.logger.WithField("gateway_id", gateway).Info("Sent downlink message")
}

// Publish implements the session.Store interface -- NON-BLOCKING
func (b *Bridge) Publish(pkt *packet.PublishPacket) {
	b.store.Publish(pkt)
	if len(pkt.TopicParts) != 2 {
		return
	}
	var msg proto.Message
	switch pkt.TopicParts[1] {
	case "up":
		msg = new(ttnv2.UplinkMessage)
	case "status":
		msg = new(ttnv2.StatusMessage)
	default:
		return
	}
	err := proto.Unmarshal(pkt.Message, msg)
	if err != nil {
		b.logger.WithField("topic", pkt.TopicName).Warn("Could not unmarshal message")
		return
	}
	if link := b.getLink(pkt.TopicParts[0]); link != nil {
		go link.SendUp(msg)
	}
}
