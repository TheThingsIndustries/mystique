// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package ttnauth implements MQTT authentication using The Things Network's account server
package ttnauth

import (
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
	logger     log.Interface
	client     *http.Client
	cache      *cache
	servers    map[string]string
	superUsers map[string]superUser
}

// SetLogger sets the logger interface.
// By default, the Noop logger is used
func (a *TTNAuth) SetLogger(logger log.Interface) {
	a.logger = logger
}

// SetCacheExpire sets the cache expiration time.
// By default, the DefaultCacheExpire is used
func (a *TTNAuth) SetCacheExpire(expire time.Duration) {
	a.cache.expire = expire
}

// AddSuperUser adds a super-user to the auth plugin
func (a *TTNAuth) AddSuperUser(username string, password []byte, access Access) {
	a.superUsers[username] = superUser{
		password: password,
		Access:   access,
	}
}

type superUser struct {
	password []byte
	Access
}

// Access information
type Access struct {
	Root         bool
	ReadPrefix   string
	Read         []string
	ReadPattern  []*regexp.Regexp
	Write        []string
	WritePattern []*regexp.Regexp
}

// IsEmpty returns true if there is no access
func (a Access) IsEmpty() bool {
	return len(a.Read)+len(a.ReadPattern)+len(a.Write)+len(a.WritePattern) == 0
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

// Connect or return error code
func (a *TTNAuth) Connect(info *auth.Info) (err error) {
	logger := a.logger.WithFields(log.F{
		"username":    info.Username,
		"remote_addr": info.RemoteAddr,
	})
	if info.RemoteHost != "" {
		logger = logger.WithField("remote_host", info.RemoteHost)
	}

	var access Access
	info.Metadata = &access
	info.Interface = a

	if superUser, ok := a.superUsers[info.Username]; ok {
		if subtle.ConstantTimeCompare(info.Password, superUser.password) != 1 {
			return packet.ConnectNotAuthorized
		}
		access = superUser.Access
		return nil
	}

	cachedAccess := a.cache.Get(info.Username, info.Password)
	if cachedAccess != nil {
		logger.Debug("Using auth result from cache")
		access = *cachedAccess
	} else {
		if !idPattern.MatchString(info.Username) {
			return packet.ConnectNotAuthorized
		}
		access.ReadPrefix = info.Username

		logger.Debug("Authenticating using account server")

		appRights, err := a.applicationRights(info.Username, string(info.Password))
		if err != nil {
			return packet.ConnectNotAuthorized
		}
		for _, right := range appRights {
			switch right {
			case "messages:up:r":
				access.ReadPattern = append(access.ReadPattern, regexp.MustCompile("^"+info.Username+"/devices/"+IDRegexp+"/up"))
				access.ReadPattern = append(access.ReadPattern, regexp.MustCompile("^"+info.Username+"/devices/"+IDRegexp+"/events"))
				access.ReadPattern = append(access.ReadPattern, regexp.MustCompile("^"+info.Username+"/events"))
			case "messages:down:w":
				access.WritePattern = append(access.WritePattern, regexp.MustCompile("^"+info.Username+"/devices/"+IDRegexp+"/down$"))
			}
		}

		gtwRights, err := a.gatewayRights(info.Username, string(info.Password))
		if err != nil {
			return packet.ConnectNotAuthorized
		}
		if len(gtwRights) > 0 {
			access.Write = append(access.Write, info.Username+"/up")
			access.Read = append(access.Read, info.Username+"/down")
			access.Write = append(access.Write, info.Username+"/status")
			access.Write = append(access.Write, "connect")
			access.Write = append(access.Write, "disconnect")
		}

		a.cache.Set(info.Username, info.Password, access)
	}

	if access.IsEmpty() {
		return packet.ConnectNotAuthorized
	}

	return nil
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
	topicParts := strings.Split(requestedTopic, topic.Separator)
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
func (a *TTNAuth) CanRead(info *auth.Info, t string) bool {
	if info.Metadata == nil {
		return false
	}
	access := info.Metadata.(*Access)
	if access.Root {
		// Root has full access
		return true
	}
	if strings.HasPrefix(t, topic.InternalPrefix) {
		// Non-root has no access to internal topics
		return false
	}
	for _, allowed := range access.Read {
		if t == allowed {
			return true
		}
	}
	for _, allowed := range access.ReadPattern {
		if allowed.MatchString(t) {
			return true
		}
	}
	return false
}

// CanWrite returns true iff the session can write to the topic
func (a *TTNAuth) CanWrite(info *auth.Info, t string) bool {
	if info.Metadata == nil {
		return false
	}
	if strings.HasPrefix(t, topic.InternalPrefix) {
		// Only the server can write to internal topics
		return false
	}
	access := info.Metadata.(*Access)
	if access.Root {
		return true
	}
	for _, allowed := range access.Write {
		if t == allowed {
			return true
		}
	}
	for _, allowed := range access.WritePattern {
		if allowed.MatchString(t) {
			return true
		}
	}
	return false
}
