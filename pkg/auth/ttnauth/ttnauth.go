// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package ttnauth implements MQTT authentication using The Things Network's account server
package ttnauth

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	"github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// IDRegexp is the regular expression that matches TTN IDs
const IDRegexp = "[0-9a-z](?:[_-]?[0-9a-z]){1,35}"

// DefaultCacheExpire sets the expiration time of the cache
var DefaultCacheExpire = time.Minute

// New returns a new auth interface that uses the TTN account server
func New(servers map[string]string) *TTNAuth {
	return &TTNAuth{
		logger:     log.Noop,
		client:     http.DefaultClient,
		cache:      newCache(DefaultCacheExpire),
		servers:    servers,
		superUsers: make(map[string]superUser),
	}
}

// TTNAuth implements authentication for TTN
type TTNAuth struct {
	logger       log.Interface
	penalty      time.Duration
	gateways     bool
	applications bool
	client       *http.Client
	cache        *cache
	servers      map[string]string
	superUsers   map[string]superUser
}

// SetLogger sets the logger interface.
// By default, the Noop logger is used
func (a *TTNAuth) SetLogger(logger log.Interface) {
	a.logger = logger
}

// SetCacheExpire sets the cache expiration time.
// By default, the DefaultCacheExpire is used
func (a *TTNAuth) SetCacheExpire(expires time.Duration) {
	a.cache.expires = expires
}

// AddSuperUser adds a super-user to the auth plugin
func (a *TTNAuth) AddSuperUser(username string, password []byte, access Access) {
	a.superUsers[username] = superUser{
		password: password,
		Access:   access,
	}
}

// SetPenalty sets the time penalty for a failed login
func (a *TTNAuth) SetPenalty(d time.Duration) {
	a.penalty = d
}

// AuthenticateGateways enables authentication of gateways
func (a *TTNAuth) AuthenticateGateways() {
	a.gateways = true
}

// AuthenticateApplications enables authentication of applications
func (a *TTNAuth) AuthenticateApplications() {
	a.applications = true
}

type superUser struct {
	password []byte
	Access
}

// Access information
type Access struct {
	Root       bool
	ReadPrefix string // must not contain `/`
	Read       [][]string
	Write      [][]string
}

// IsEmpty returns true if there is no access
func (a Access) IsEmpty() bool {
	return len(a.Read)+len(a.Write) == 0
}

var idPattern = regexp.MustCompile("^[0-9a-z](?:[_-]?[0-9a-z]){1,35}$")

func (a *TTNAuth) rights(entity, id, key string) (rights []string, err error) {
	keyParts := strings.SplitN(key, ".", 2)
	if len(keyParts) != 2 {
		return nil, errors.New("invalid access key")
	}
	server, ok := a.servers[keyParts[0]]
	if !ok {
		return nil, fmt.Errorf("identity server %s not found", keyParts[0])
	}
	req, err := http.NewRequest("GET", server+fmt.Sprintf("/api/v2/%s/%s/rights", entity, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Key "+key)
	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()
	if res.StatusCode != 200 {
		return nil, nil
	}
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&rights)
	return
}

func (a *TTNAuth) gatewayRights(gatewayID, key string) ([]string, error) {
	return a.rights("gateways", gatewayID, key)
}

func (a *TTNAuth) applicationRights(applicationID, key string) ([]string, error) {
	return a.rights("applications", applicationID, key)
}

func (a *TTNAuth) FetchAccess(username string, password []byte) (*Access, error) {
	access := Access{}

	if !idPattern.MatchString(username) {
		return nil, packet.ConnectNotAuthorized
	}
	access.ReadPrefix = username

	if a.applications {
		appRights, err := a.applicationRights(username, string(password))
		if err != nil {
			return nil, packet.ConnectNotAuthorized
		}
		for _, right := range appRights {
			switch right {
			case "messages:up:r":
				access.Read = append(access.Read, []string{username, "devices", topic.PartWildcard, "up"})
				access.Read = append(access.Read, []string{username, "devices", topic.PartWildcard, "up", topic.Wildcard})
				access.Read = append(access.Read, []string{username, "devices", topic.PartWildcard, "events"})
				access.Read = append(access.Read, []string{username, "devices", topic.PartWildcard, "events", topic.Wildcard})
				access.Read = append(access.Read, []string{username, "events"})
				access.Read = append(access.Read, []string{username, "events", topic.Wildcard})
			case "messages:down:w":
				access.Write = append(access.Write, []string{username, "devices", "+", "down"})
			}
		}
	}

	if a.gateways {
		gtwRights, err := a.gatewayRights(username, string(password))
		if err != nil {
			return nil, packet.ConnectNotAuthorized
		}
		if len(gtwRights) > 0 {
			access.Write = append(access.Write, []string{username, "up"})
			access.Read = append(access.Read, []string{username, "down"})
			access.Write = append(access.Write, []string{username, "status"})
			access.Write = append(access.Write, []string{"connect"})
			access.Write = append(access.Write, []string{"disconnect"})
		}
	}

	return &access, nil
}

// Connect or return error code
func (a *TTNAuth) Connect(ctx context.Context, info *auth.Info) (context.Context, error) {
	var err error
	if a.penalty > 0 {
		defer func() {
			if err != nil {
				time.Sleep(a.penalty)
			}
		}()
	}

	var access Access
	info.Metadata = &access
	info.Interface = a

	if superUser, ok := a.superUsers[info.Username]; ok {
		if subtle.ConstantTimeCompare(info.Password, superUser.password) != 1 {
			return nil, packet.ConnectNotAuthorized
		}
		access = superUser.Access
		return ctx, nil
	}

	cachedAccess, err := a.cache.GetOrFetch(info.Username, info.Password, a.FetchAccess)
	if err != nil {
		return nil, err
	}
	access = *cachedAccess

	if access.IsEmpty() {
		return nil, packet.ConnectNotAuthorized
	}

	return ctx, nil
}

// RouterAccess gives the access rights for a Router
var RouterAccess = Access{
	Read: [][]string{
		{"connect"},
		{"disconnect"},
		{topic.PartWildcard, "up"},
		{topic.PartWildcard, "status"},
	},
	Write: [][]string{
		{topic.PartWildcard, "down"},
	},
}

// HandlerAccess gives the access rights for a Handler
var HandlerAccess = Access{
	Read: [][]string{
		{topic.PartWildcard, "devices", topic.PartWildcard, "down"},
	},
	Write: [][]string{
		{topic.PartWildcard, "devices", topic.PartWildcard, "up"},
		{topic.PartWildcard, "devices", topic.PartWildcard, "up", topic.Wildcard},
		{topic.PartWildcard, "devices", topic.PartWildcard, "events"},
		{topic.PartWildcard, "devices", topic.PartWildcard, "events", topic.Wildcard},
		{topic.PartWildcard, "events"},
		{topic.PartWildcard, "events", topic.Wildcard},
	},
}

// Subscribe allows the auth plugin to replace wildcards or to lower the QoS of a subscription.
// For example, a client requesting a subscription to "#" may be rewritten to "foo/#" if they are only allowed to subscribe to that topic.
func (a *TTNAuth) Subscribe(info *auth.Info, requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error) {
	if info.Metadata == nil {
		return acceptedTopic, acceptedQoS, errors.New("No auth metadata present")
	}
	acceptedTopic = requestedTopic
	acceptedQoS = requestedQoS
	access := info.Metadata.(*Access)
	if access.Root {
		return
	}
	if access.ReadPrefix == "" {
		return
	}
	topicParts := topic.Split(requestedTopic)
	switch topicParts[0] {
	case topic.Wildcard:
		acceptedTopic = access.ReadPrefix + "/#"
	case topic.PartWildcard:
		topicParts[0] = access.ReadPrefix
		acceptedTopic = strings.Join(topicParts, topic.Separator)
	case access.ReadPrefix:
	default:
		err = errors.New("not authorized on this topic")
	}
	return
}

// CanRead returns true iff the session can read from the topic
func (a *TTNAuth) CanRead(info *auth.Info, t ...string) bool {
	switch len(t) {
	case 0:
		return false
	case 1:
		t = topic.Split(t[0])
	}
	access, ok := info.Metadata.(*Access)
	if !ok {
		return false
	}
	if access.Root {
		// Root has full access
		return true
	}
	if access.ReadPrefix != "" && t[0] != access.ReadPrefix {
		// No access if prefix is set and does not match
		return false
	}
	for _, allowed := range access.Read {
		if topic.MatchPath(t, allowed) {
			return true
		}
	}
	return false
}

// CanWrite returns true iff the session can write to the topic
func (a *TTNAuth) CanWrite(info *auth.Info, t ...string) bool {
	switch len(t) {
	case 0:
		return false
	case 1:
		t = topic.Split(t[0])
	}
	access, ok := info.Metadata.(*Access)
	if !ok {
		return false
	}
	if access.Root {
		if strings.HasPrefix(t[0], topic.InternalPrefix) {
			// Only the server can write to internal topics
			return false
		}
		return true
	}
	for _, allowed := range access.Write {
		if topic.MatchPath(t, allowed) {
			return true
		}
	}
	return false
}
