// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package subscription implements MQTT topic subscriptions
package subscription

import (
	"sync"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

type subscription struct {
	filter     string
	filterPath []string
	qos        byte
}

func (s subscription) Match(topicPath []string) bool {
	return topic.MatchPath(topicPath, s.filterPath)
}

// List of MQTT topic subscriptions
type List struct {
	mu            sync.RWMutex
	subscriptions []subscription
}

// Add a subscription to the list
func (s *List) Add(filter string, qos byte) (added bool) {
	if len(filter) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sub := subscription{
		filter:     filter,
		filterPath: topic.Split(filter),
		qos:        qos,
	}
	for i, existing := range s.subscriptions {
		if existing.filter == filter {
			s.subscriptions[i] = sub
			return
		}
	}
	s.subscriptions = append(s.subscriptions, sub)
	subscriptionsGauge.Inc()
	return true
}

// Remove a subscription from the list
func (s *List) Remove(filter string) (removed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sub := range s.subscriptions {
		if sub.filter == filter {
			s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)
			removed = true
			subscriptionsGauge.Dec()
			return
		}
	}
	return
}

// Clear the subscription list
func (s *List) Clear() {
	s.mu.Lock()
	subscriptionsGauge.Sub(float64(len(s.subscriptions)))
	s.subscriptions = nil
	s.mu.Unlock()
}

// Match the topic to the subscriptions and return the maximum QoS
func (s *List) Match(t ...string) (qos byte, found bool) {
	switch len(t) {
	case 0:
		return
	case 1:
		t = topic.Split(t[0])
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.subscriptions {
		if sub.Match(t) {
			found = true
			if sub.qos > qos {
				qos = sub.qos
			}
		}
	}
	return
}

// Matches for the topic
func (s *List) Matches(t ...string) (matches []string) {
	switch len(t) {
	case 0:
		return
	case 1:
		t = topic.Split(t[0])
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.subscriptions {
		if sub.Match(t) {
			matches = append(matches, sub.filter)
		}
	}
	return
}

// Count the subscriptions
func (s *List) Count() (count int) {
	s.mu.RLock()
	count = len(s.subscriptions)
	s.mu.RUnlock()
	return
}

// Subscriptions returns the subscriptions in the list
func (s *List) Subscriptions() map[string]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	subscriptions := make(map[string]byte, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		subscriptions[sub.filter] = sub.qos
	}
	return subscriptions
}
