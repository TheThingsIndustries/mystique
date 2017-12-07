// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package auth defines the authentication interface for MQTT.
package auth

import (
	"errors"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

// Interface for MQTT authentication
type Interface interface {
	// Connect or return error code
	Connect(info *Info) error

	// Subscribe allows the auth plugin to replace wildcards or to lower the QoS of a subscription.
	// For example, a client requesting a subscription to "#" may be rewritten to "foo" if they are only allowed to subscribe to that topic.
	Subscribe(info *Info, requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error)

	// Can the session read from the (application-layer) topic
	CanRead(info *Info, topic ...string) bool

	// Can the session write to the (application-layer) topic
	CanWrite(info *Info, topic ...string) bool
}

// Info for an MQTT user
type Info struct {
	Interface
	RemoteAddr string
	RemoteHost string
	Transport  string
	ClientID   string
	Username   string
	Password   []byte
	Metadata   interface{}
}

// Subscribe to the requested topic and QoS, which can be adapted by the auth plugin
func (i *Info) Subscribe(requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error) {
	if i == nil {
		return requestedTopic, requestedQoS, errors.New("no auth info present")
	}
	if i.Interface == nil {
		return requestedTopic, requestedQoS, nil
	}
	return i.Interface.Subscribe(i, requestedTopic, requestedQoS)
}

// CanRead returns true iff given the info, the client can read on a topic
func (i *Info) CanRead(t ...string) bool {
	if len(t) == 1 {
		t = topic.Split(t[0])
	}
	if i == nil {
		return false // won't allow that if there's no auth info
	}
	if i.Interface == nil {
		return true
	}
	return i.Interface.CanRead(i, t...)
}

// CanWrite returns true iff given the info, the client can write on a topic
func (i *Info) CanWrite(t ...string) bool {
	if len(t) == 1 {
		t = topic.Split(t[0])
	}
	if i == nil {
		return false // won't allow that if there's no auth info
	}
	if i.Interface == nil {
		return true
	}
	return i.Interface.CanWrite(i, t...)
}
